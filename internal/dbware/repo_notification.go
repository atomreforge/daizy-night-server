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

func (r *RepoNotification) CreateItem(n model.Notification) error {
	return r.pDB.DB().Transaction(func(tx *gorm.DB) error {
		if tx.Create(n).Error != nil {
			if errors.Is(tx.Error, gorm.ErrDuplicatedKey) {
				return nil
			}
			return tx.Error
		}
		return nil
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
				http.StatusInternalServerError,
				string(consts.ExprNotificationItem),
			)
		}
		return model.Notification{}, res.Error
	}
	return n, nil
}

// get latest nid in repo
func (r *RepoNotification) GetLatestNotificationNid() (uint, error) {
	var n model.Notification
	res := r.pDB.DB().Order("nid desc").First(&n)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return 0, errs.BuildErrDbRecord(
				errs.DbRecordNotFound,
				http.StatusInternalServerError,
				string(consts.ExprNotificationItem),
			)
		}
		return 0, res.Error
	}
	return n.Nid, nil
}
