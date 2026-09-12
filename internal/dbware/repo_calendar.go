package dbware

import (
	"errors"
	"net/http"

	abstract "github.com/atomreforge/daizy-night-server/internal/abstract/interface"
	"github.com/atomreforge/daizy-night-server/internal/consts"
	"github.com/atomreforge/daizy-night-server/internal/errs"
	"github.com/atomreforge/daizy-night-server/internal/model"
	"github.com/atomreforge/daizy-night-server/internal/utils"

	"gorm.io/gorm"
)

var _ abstract.InterfaceRepoCalendar = (*RepoCalendar)(nil)

// maxGenIDAttempts bounds the retry loop for the rare collision of the
// random 7-digit business id.
const maxGenIDAttempts = 5

type RepoCalendar struct {
	pDB abstract.InterfaceProviderDB
}

func NewRepoCalendar(pDB abstract.InterfaceProviderDB) *RepoCalendar {
	return &RepoCalendar{pDB: pDB}
}

func (r *RepoCalendar) RecordNewCalendar(b *model.CalendarTable) error {
	return r.pDB.DB().Transaction(func(tx *gorm.DB) error {
		// single calendar per user
		var count int64
		if err := tx.Model(&model.CalendarTable{}).Where("user_id = ?", b.UserID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errs.BuildErrValidation(errs.ValidationKeyDuplicatedValue, http.StatusBadRequest, string(consts.ExprUserID), string(consts.ExprCalendar))
		}
		return createCalendar(tx, b)
	})
}

func (r *RepoCalendar) GetCalendarByUid(uid uint) (*model.CalendarTable, error) {
	var rec model.CalendarTable
	res := r.pDB.DB().
		Preload("Records", func(db *gorm.DB) *gorm.DB { return db.Order("weekday ASC, start_min ASC") }).
		Where("user_id = ?", uid).First(&rec)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, errs.BuildErrDbRecord(errs.DbRecordNotFound, http.StatusNotFound, string(consts.ExprCalendar))
		}
		return nil, res.Error
	}
	return &rec, nil
}

// UpdateCalendar performs a full replacement of the user's timetable and
// auto looks up the row by uid; the row is created on first use, so a PUT
// never has to branch between create and update.
func (r *RepoCalendar) UpdateCalendar(b *model.CalendarTable) error {
	return r.pDB.DB().Transaction(func(tx *gorm.DB) error {
		var cur model.CalendarTable
		err := tx.Where("user_id = ?", b.UserID).First(&cur).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return createCalendar(tx, b)
		}
		if err != nil {
			return err
		}

		// swap the stored items for b.Records
		if err := tx.Where("calendar_id = ?", cur.CalendarID).Delete(&model.CalendarItem{}).Error; err != nil {
			return err
		}
		items := make([]model.CalendarItem, 0, len(b.Records))
		for _, it := range b.Records {
			it.ID = 0 // never reuse soft-deleted primary keys
			it.CalendarID = cur.CalendarID
			items = append(items, it)
		}
		if len(items) > 0 {
			return tx.Create(&items).Error
		}
		return nil
	})
}

// RemoveCalendarByUid removes the user's calendar together with its items
// (soft delete; the sqlite foreign-key pragma is off, so a db-level
// CASCADE would never fire). idempotent: succeeds when none exists.
func (r *RepoCalendar) RemoveCalendarByUid(uid uint) error {
	return r.pDB.DB().Transaction(func(tx *gorm.DB) error {
		var cur model.CalendarTable
		err := tx.Where("user_id = ?", uid).First(&cur).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := tx.Where("calendar_id = ?", cur.CalendarID).Delete(&model.CalendarItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&cur).Error
	})
}

// RemoveCalendarByModel removes by primary key, cascading the items.
func (r *RepoCalendar) RemoveCalendarByModel(b *model.CalendarTable) error {
	if b == nil || b.ID == 0 {
		return errs.BuildErrDbRecord(errs.DbRecordNotFound, http.StatusBadRequest, string(consts.ExprCalendar))
	}
	return r.pDB.DB().Transaction(func(tx *gorm.DB) error {
		var cur model.CalendarTable
		if err := tx.First(&cur, b.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		if err := tx.Where("calendar_id = ?", cur.CalendarID).Delete(&model.CalendarItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&cur).Error
	})
}

// createCalendar generates the business id (same recipe as CreateUser) and
// inserts the calendar row plus its items.
func createCalendar(tx *gorm.DB, b *model.CalendarTable) error {
	for i := 0; i < maxGenIDAttempts; i++ {
		id, err := utils.GenUid()
		if err != nil {
			return err
		}
		b.CalendarID = id

		if err := tx.Create(b).Error; err == nil {
			return nil
		} else if !errors.Is(err, gorm.ErrDuplicatedKey) {
			return err
		}
	}
	return errs.BuildErrValidation(errs.ValidationKeyDuplicatedValue, http.StatusInternalServerError, "calendar_id", "")
}
