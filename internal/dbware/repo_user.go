package dbware

import (
	"errors"
	"net/http"
	"time"

	abstract "github.com/atomreforge/daizy-night-server/internal/abstract/interface"
	"github.com/atomreforge/daizy-night-server/internal/consts"
	"github.com/atomreforge/daizy-night-server/internal/errs"
	"github.com/atomreforge/daizy-night-server/internal/model"
	"github.com/atomreforge/daizy-night-server/internal/utils"

	"gorm.io/gorm"
)

var _ abstract.InterfaceRepoUser = (*RepoUser)(nil)

type RepoUser struct {
	pDB abstract.InterfaceProviderDB
}

func NewRepoUser(pDB abstract.InterfaceProviderDB) *RepoUser {
	return &RepoUser{pDB: pDB}
}

// CreateUser shares the maxGenIDAttempts retry bound declared in
// repo_calendar.go for the same random 7-digit business id recipe.
func (r *RepoUser) CreateUser(b *model.User) error {
	return r.pDB.DB().Transaction(func(tx *gorm.DB) error {
		// precise attribution pre-checks. the sqlite driver translates every
		// unique violation into gorm.ErrDuplicatedKey without revealing which
		// index fired, so conflicts are identified by cheap indexed lookups.
		var n int64
		if err := tx.Model(&model.User{}).Where("username = ?", b.Username).Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			return errs.BuildErrValidation(errs.ValidationKeyDuplicatedValue, http.StatusBadRequest, string(consts.UiExprUsername), b.Username)
		}
		if err := tx.Model(&model.User{}).Where("registercode = ?", string(b.Registercode)).Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			return errs.BuildErrRegistercode(errs.RegistercodeUnusableUsed, http.StatusBadRequest)
		}

		for i := 0; i < maxGenIDAttempts; i++ {
			uid, err := utils.GenRandID()
			if err != nil {
				return err
			}

			// cover attrs
			b.UserID = uid
			b.Role = consts.User
			b.RegisterTime = time.Now()

			// formal repo ops
			err = tx.Create(b).Error
			if err == nil {
				return claimRegistercode(tx, b)
			}
			if !errors.Is(err, gorm.ErrDuplicatedKey) {
				return err
			}

			// a concurrent registration may have taken the username or the
			// registercode after the pre-checks; re-check to attribute, and
			// otherwise treat the violation as a user_id collision.
			if err := attributeDuplicate(tx, b); err != nil {
				return err
			}
		}
		return errs.BuildErrValidation(errs.ValidationKeyDuplicatedValue, http.StatusInternalServerError, "uid", "")
	})
}

// attributeDuplicate re-checks which unique column caused a violation right
// after gorm.ErrDuplicatedKey. It returns a precise business error when the
// conflict is on username or registercode, and nil when the remaining
// explanation is a user_id collision (the caller retries with a new uid).
func attributeDuplicate(tx *gorm.DB, b *model.User) error {
	var n int64
	if err := tx.Model(&model.User{}).Where("username = ?", b.Username).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return errs.BuildErrValidation(errs.ValidationKeyDuplicatedValue, http.StatusBadRequest, string(consts.UiExprUsername), b.Username)
	}
	if err := tx.Model(&model.User{}).Where("registercode = ?", string(b.Registercode)).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return errs.BuildErrRegistercode(errs.RegistercodeUnusableUsed, http.StatusBadRequest)
	}
	return nil
}

// claimRegistercode marks the registercode row as used inside the same
// transaction that created the user. A zero-row update means the code was
// claimed concurrently and rolls the whole registration back.
func claimRegistercode(tx *gorm.DB, b *model.User) error {
	res := tx.Model(&model.RegistercodeRecord{}).
		Where("raw_hex = ? AND used = ?", string(b.Registercode), false).
		Updates(map[string]any{"used": true, "user_id": b.UserID})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return errs.BuildErrRegistercode(errs.RegistercodeUnusableUsed, http.StatusBadRequest)
	}
	return nil
}

// Repo contract: both getters return a business error on "not found"
// (GetUserByUsername → 400 for the login flow's generic mapping;
// GetUserByUid → 404 per the API docs). (nil, nil) is never returned.
func (r *RepoUser) GetUserByUsername(name string) (*model.User, error) {
	var user model.User
	result := r.pDB.DB().Where(&model.User{Username: name}).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errs.BuildErrDbRecord(errs.DbRecordUsernameNotFound, http.StatusBadRequest, string(consts.InlineExprUser))
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *RepoUser) GetUserByUid(uid uint) (*model.User, error) {
	var user model.User
	result := r.pDB.DB().Where("user_id = ?", uid).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errs.BuildErrDbRecord(errs.DbRecordNotFound, http.StatusNotFound, string(consts.InlineExprUser))
		}
		return nil, result.Error
	}
	return &user, nil
}
