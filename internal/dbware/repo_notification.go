package dbware

import (
	"errors"
	"net/http"

	abstract "github.com/atomreforge/daizy-night-server/internal/abstract/interface"
	"github.com/atomreforge/daizy-night-server/internal/consts"
	"github.com/atomreforge/daizy-night-server/internal/errs"
	"github.com/atomreforge/daizy-night-server/internal/model"
	"gorm.io/gorm"
)

var _ abstract.InterfaceRepoNotification = (*RepoNotification)(nil)

type RepoNotification struct {
	pDB abstract.InterfaceProviderDB
}

func NewRepoNotification(pDB abstract.InterfaceProviderDB) *RepoNotification {
	return &RepoNotification{pDB: pDB}
}

// CreateItem inserts a notification. `Nid` is always server-allocated: a
// strictly positive, monotonically increasing integer (COALESCE(MAX(nid),0)+1,
// taken over every row — soft-deleted included — so the nid of a removed
// notification is never reused). Nid starts at 1 and is never zero; any
// caller-provided `Nid` is ignored and overwritten. SQLite serializes writes,
// so the in-transaction MAX+1 allocation is race-free.
func (r *RepoNotification) CreateItem(n model.Notification) error {
	return r.pDB.DB().Transaction(func(tx *gorm.DB) error {
		// next nid
		var nextNid uint
		if err := tx.Raw(`SELECT COALESCE(MAX(nid), 0) + 1 FROM notifications`).Scan(&nextNid).Error; err != nil {
			return err
		}
		n.Nid = nextNid

		// repo ops
		return tx.Create(&n).Error
	})
}

func (r *RepoNotification) RemoveItem(nid uint) error {
	return r.pDB.DB().Transaction(func(tx *gorm.DB) error {
		return tx.Where("nid = ?", nid).
			Delete(&model.Notification{}).Error
	})
}

func (r *RepoNotification) UpdateItem(n model.Notification) error {
	rec, err := r.GetItemByNid(n.Nid)
	if err != nil {
		return err
	}
	n.ID = rec.ID
	// preserve server-owned columns: zero values coming from the caller must
	// not overwrite the persisted ones on Save()
	n.CreatedAt = rec.CreatedAt
	n.DeletedAt = rec.DeletedAt
	n.PostBy = rec.PostBy

	return r.pDB.DB().Transaction(func(tx *gorm.DB) error {
		return tx.Model(&rec).Save(&n).Error
	})
}

func (r *RepoNotification) GetItemByNid(nid uint) (model.Notification, error) {
	var n model.Notification
	res := r.pDB.DB().Where("nid = ?", nid).First(&n)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return model.Notification{}, errs.BuildErrDbRecord(
				errs.DbRecordNotFound,
				http.StatusNotFound,
				string(consts.ExprNotificationItem),
			)
		}
		return model.Notification{}, res.Error
	}
	return n, nil
}

// get latest `nid` in repo; with monotonic nid allocation the greatest live
// nid is simply the greatest one (soft-deleted rows are excluded by gorm)
func (r *RepoNotification) GetLatestNotificationNid() (uint, error) {
	var n model.Notification
	res := r.pDB.DB().Order("nid DESC").First(&n)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return 0, nil
		} else {
			return 0, errs.BuildErrDbRecord(
				errs.DbRecordNotFound,
				http.StatusInternalServerError,
				string(consts.ExprNotificationItem))
		}
	}
	return n.Nid, nil
}

// GetNotificationBatch fetches notifications whose `nid` lies in the closed
// range [nidBegin, nidEnd], ascending by nid. The returned slice is never
// nil, so an empty result marshals to `[]` rather than `null`.
func (r *RepoNotification) GetNotificationBatch(nidBegin, nidEnd uint) (*[]model.Notification, error) {
	batch := []model.Notification{}

	res := r.pDB.DB().
		Where("nid >= ? AND nid <= ?", nidBegin, nidEnd).
		Order("nid ASC").
		Find(&batch)
	if res.Error != nil {
		return nil, res.Error
	}
	return &batch, nil
}
