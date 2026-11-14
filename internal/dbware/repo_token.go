package dbware

import (
	"errors"
	"net/http"
	"time"

	abstract "github.com/atomreforge/daizy-night-server/internal/abstract/interface"
	"github.com/atomreforge/daizy-night-server/internal/config"
	"github.com/atomreforge/daizy-night-server/internal/consts"
	"github.com/atomreforge/daizy-night-server/internal/errs"
	"github.com/atomreforge/daizy-night-server/internal/model"
	"github.com/atomreforge/daizy-night-server/internal/utils"
	"gorm.io/gorm"
)

var _ abstract.InterfaceRepoToken = (*RepoToken)(nil)

type RepoToken struct {
	pDB abstract.InterfaceProviderDB
	cfg *config.Config
}

func NewRepoToken(pDB abstract.InterfaceProviderDB, cfg *config.Config) *RepoToken {
	return &RepoToken{pDB: pDB, cfg: cfg}
}

func (r *RepoToken) SaveRefreshToken(uid uint, rawToken string) error {
	// prune tokens that can never authenticate again
	if err := r.PruneStaleTokens(uid); err != nil {
		return err
	}

	// only the SHA256 lookup hash is stored: the raw token never touches
	// the database, and the expensive argon2 hash of the token proved to be
	// write-only dead weight (nothing ever compared TokenHash).
	return r.pDB.DB().Create(&model.RefreshToken{
		Uid:        uid,
		LookupHash: utils.SHA256HashHex(rawToken),
	}).Error
}

func (r *RepoToken) GetRefreshToken(uid uint, rawToken string) (*model.RefreshToken, error) {
	var token model.RefreshToken
	err := r.pDB.DB().
		Where("uid = ? AND lookup_hash = ? AND revoked_at IS NULL", uid, utils.SHA256HashHex(rawToken)).
		First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

func (r *RepoToken) RevokeUserTokens(uid uint) error {
	now := time.Now()
	return r.pDB.DB().Model(&model.RefreshToken{}).
		Where("uid = ? AND revoked_at IS NULL", uid).
		Update("revoked_at", now).Error
}

// RevokeRefreshToken revokes a single refresh token of a user. Tokens that are
// unknown or already revoked affect zero rows: signout stays idempotent.
func (r *RepoToken) RevokeRefreshToken(uid uint, rawToken string) error {
	return r.pDB.DB().Model(&model.RefreshToken{}).
		Where("uid = ? AND lookup_hash = ? AND revoked_at IS NULL", uid, utils.SHA256HashHex(rawToken)).
		Update("revoked_at", time.Now()).Error
}

// pruneStale removes rows of a user that can never authenticate again:
//   - revoked, and past the revoked-tokens retention window;
//   - expired-and-stale: older than the refresh lifetime plus the retention
//     buffer, unless the revocation itself is still inside the retention
//     window (documented semantics: a revoked record is kept for 72h).
//
// Before the second rule, tokens that expired naturally (revoked_at IS NULL)
// were never cleaned up and the table grew without bound. It runs inside the
// caller's transaction when invoked from RotateRefreshToken.
func (r *RepoToken) pruneStale(tx *gorm.DB, uid uint) error {
	now := time.Now()
	cutoffRevoked := now.Add(-r.cfg.Security.JwtRevokedTokensRetainTime)
	if err := tx.Where(
		"uid = ? AND revoked_at IS NOT NULL AND revoked_at < ?",
		uid, cutoffRevoked,
	).Delete(&model.RefreshToken{}).Error; err != nil {
		return err
	}
	cutoffExpired := now.Add(-(r.cfg.Security.JwtRefreshTokenExpireTime + r.cfg.Security.JwtRevokedTokensRetainTime))
	return tx.Where(
		"uid = ? AND created_at < ? AND (revoked_at IS NULL OR revoked_at < ?)",
		uid, cutoffExpired, cutoffRevoked,
	).Delete(&model.RefreshToken{}).Error
}

// PruneStaleTokens prunes unusable tokens of a user on the shared connection.
func (r *RepoToken) PruneStaleTokens(uid uint) error {
	return r.pruneStale(r.pDB.DB(), uid)
}

func (r *RepoToken) RotateRefreshToken(uid uint, usedLookupHash string, newRawToken string) error {
	newLookupHash := utils.SHA256HashHex(newRawToken)

	return r.pDB.DB().Transaction(func(tx *gorm.DB) error {
		if err := r.pruneStale(tx, uid); err != nil {
			return err
		}

		res := tx.Model(&model.RefreshToken{}).
			Where("uid = ? AND lookup_hash = ? AND revoked_at IS NULL", uid, usedLookupHash).
			Update("revoked_at", time.Now())
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errs.BuildErrJwtToken(
				errs.JwtRefreshTokenUsed,
				http.StatusUnauthorized,
			)
		}

		return tx.Create(&model.RefreshToken{
			Uid:        uid,
			LookupHash: newLookupHash,
			RevokedAt:  nil,
		}).Error
	})
}

func (r *RepoToken) GetUidByRefreshToken(rawToken string) (uint, error) {
	var token model.RefreshToken
	res := r.pDB.DB().
		Where("lookup_hash = ? AND revoked_at IS NULL", utils.SHA256HashHex(rawToken)).
		First(&token)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return 0,
				errs.BuildErrUserLogin(
					errs.UserLoginParamsIncorrect,
					http.StatusUnauthorized,
					string(consts.ExprIndetermined),
				)
		}
		return 0, res.Error
	}
	return token.Uid, nil
}

func (r *RepoToken) GetUserByRefreshToken(rawToken string) (*model.User, error) {
	var token model.RefreshToken
	res := r.pDB.DB().
		Where("lookup_hash = ? AND revoked_at IS NULL", utils.SHA256HashHex(rawToken)).
		First(&token)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil,
				errs.BuildErrDbRecord(
					errs.DbRecordNotFound,
					http.StatusBadRequest,
					string(consts.HttpExprErrorBadRequest),
				)
		}
		return nil, res.Error
	}

	var user model.User
	res = r.pDB.DB().Where("user_id = ?", token.Uid).First(&user)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil,
				errs.BuildErrDbRecord(
					errs.DbRecordNotFound,
					http.StatusNotFound,
					string(consts.InlineExprUser),
				)
		}
		return nil, res.Error
	}
	return &user, nil
}
