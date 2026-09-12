package dbware

import (
	"errors"
	"net/http"

	"gorm.io/gorm"

	abstract "github.com/atomreforge/daizy-night-server/internal/abstract/interface"
	"github.com/atomreforge/daizy-night-server/internal/consts"
	"github.com/atomreforge/daizy-night-server/internal/errs"
	"github.com/atomreforge/daizy-night-server/internal/model"
)

var _ abstract.InterfaceRepoRegistercode = (*RepoRegistercode)(nil)

type RepoRegistercode struct {
	pDB abstract.InterfaceProviderDB
}

func NewRepoRegistercode(pDB abstract.InterfaceProviderDB) *RepoRegistercode {
	return &RepoRegistercode{pDB: pDB}
}

// once the server contacts a registercode(no matter whether valid ot not)
// it records it.
// [!] notice that the feature is possible to be exploited by attackers if any bcz the register process doesnt req auth.
func (r *RepoRegistercode) RecordNewRegistercode(b *model.RegistercodeRecord) error {
	return r.pDB.DB().Transaction(func(tx *gorm.DB) error {
		if tx.Create(b).Error != nil {
			if errors.Is(tx.Error, gorm.ErrDuplicatedKey) {
				return nil
			}
			return tx.Error
		}
		return nil
	})
}

func (r *RepoRegistercode) Remove(rawHex model.RegistercodeRawHex) error {
	return r.pDB.DB().Transaction(func(tx *gorm.DB) error {
		return tx.Where("raw_hex = ?", string(rawHex)).
			Delete(&model.RegistercodeRecord{}).Error
	})
}

// mark an exsiting registercode as used by a user;
// the metaphor is that the user has "consumed" the registercode, so it cannot be used again.
func (r *RepoRegistercode) Used(rawHex model.RegistercodeRawHex, value bool, userID uint) error {
	rec, err := r.GetRecordByRegistercode(rawHex)
	if err != nil {
		return err
	}
	return r.pDB.DB().Transaction(func(tx *gorm.DB) error {
		return tx.Model(&rec).Updates(map[string]any{
			"used":    value,
			"user_id": userID,
		}).Error
	})
}

// fully update.
func (r *RepoRegistercode) Updates(record model.RegistercodeRecord) error {
	rec, err := r.GetRecordByRegistercode(record.RawHex)
	if err != nil {
		return err
	}
	record.ID = rec.ID

	return r.pDB.DB().Transaction(func(tx *gorm.DB) error {
		return tx.Model(&rec).Save(&record).Error
	})
}

func (r *RepoRegistercode) GetRecordByRegistercode(rawHex model.RegistercodeRawHex) (*model.RegistercodeRecord, error) {
	var rec model.RegistercodeRecord
	result := r.pDB.DB().Where("raw_hex = ?", string(rawHex)).First(&rec)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errs.BuildErrDbRecord(errs.DbRecordNotFound, http.StatusInternalServerError, string(consts.ExprRegistercodeRecord))
		}
		return nil, result.Error
	}
	return &rec, nil
}
