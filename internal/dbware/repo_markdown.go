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

var _ abstract.InterfaceRepoMarkdown = (*RepoMarkdown)(nil)

type RepoMarkdown struct {
	pDB abstract.InterfaceProviderDB
}

func NewRepoMarkdown(pDB abstract.InterfaceProviderDB) *RepoMarkdown {
	return &RepoMarkdown{pDB: pDB}
}

func (r *RepoMarkdown) CreateItem(md model.Markdown) error {
	return r.pDB.DB().Transaction(func(tx *gorm.DB) error {
		if tx.Create(md).Error != nil {
			if errors.Is(tx.Error, gorm.ErrDuplicatedKey) {
				return nil
			}
			return tx.Error
		}
		return nil
	})
}

func (r *RepoMarkdown) RemoveItem(mdid uint) error {
	return r.pDB.DB().Transaction(func(tx *gorm.DB) error {
		return tx.Where("mdid = ?", mdid).
			Delete(&model.Markdown{}).Error
	})
}

func (r *RepoMarkdown) Updates(md model.Markdown) error {
	rec, err := r.GetMdByMdid(md.Mdid)
	if err != nil {
		return err
	}
	md.ID = rec.ID

	return r.pDB.DB().Transaction(func(tx *gorm.DB) error {
		return tx.Model(&rec).Save(&md).Error
	})
}

func (r *RepoMarkdown) GetMdByMdid(mdid uint) (model.Markdown, error) {
	var md model.Markdown
	res := r.pDB.DB().Where("mdid = ?", mdid).First(&md)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return model.Markdown{}, errs.BuildErrDbRecord(
				errs.DbRecordNotFound,
				http.StatusInternalServerError,
				string(consts.ExprMarkdownItem),
			)
		}
		return model.Markdown{}, res.Error
	}
	return md, nil
}
