package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"

	abstract "github.com/atomreforge/daizy-night-server/internal/abstract/interface"
	"github.com/atomreforge/daizy-night-server/internal/config"
	"github.com/atomreforge/daizy-night-server/internal/consts"
	"github.com/atomreforge/daizy-night-server/internal/errs"
	"github.com/atomreforge/daizy-night-server/internal/model"
	"github.com/atomreforge/daizy-night-server/internal/utils"

	"time"

	"github.com/golang-jwt/jwt/v5"
)

var _ abstract.InterfaceCrypto = (*ProviderCrypto)(nil)

type ProviderCrypto struct {
	JwtAccessTokenEnckey   ed25519.PrivateKey
	JwtAccessTokenDeckey   ed25519.PublicKey
	JwtRefreshTokenEnckey  ed25519.PrivateKey
	JwtRefreshTokenDeckey  ed25519.PublicKey
	RegistercodeEnckey     ed25519.PrivateKey
	RegistercodeDeckey     ed25519.PublicKey
	AccessTokenExpireTime  time.Duration
	RefreshTokenExpireTime time.Duration
}

// summon a new crypto provider.
func NewProviderCrypto(cfg *config.Config) (*ProviderCrypto, error) {
	var errs []error
	jate, err := utils.HexToPrivKey(cfg.Security.JwtAccessTokenEnckey)
	errs = append(errs, err)
	jatd, err := utils.HexToPubKey(cfg.Security.JwtAccessTokenDeckey)
	errs = append(errs, err)
	jrte, err := utils.HexToPrivKey(cfg.Security.JwtRefreshTokenEnckey)
	errs = append(errs, err)
	jrtd, err := utils.HexToPubKey(cfg.Security.JwtRefreshTokenDeckey)
	errs = append(errs, err)
	re, err := utils.HexToPrivKey(cfg.Security.RegistercodeEnckey)
	errs = append(errs, err)
	rd, err := utils.HexToPubKey(cfg.Security.RegistercodeDeckey)
	errs = append(errs, err)
	atet := cfg.Security.JwtAccessTokenExpireTime
	rtet := cfg.Security.JwtRefreshTokenExpireTime

	if err := errors.Join(errs...); err != nil {
		return nil, err
	}

	return &ProviderCrypto{
		JwtAccessTokenEnckey:   jate,
		JwtAccessTokenDeckey:   jatd,
		JwtRefreshTokenEnckey:  jrte,
		JwtRefreshTokenDeckey:  jrtd,
		RegistercodeEnckey:     re,
		RegistercodeDeckey:     rd,
		AccessTokenExpireTime:  atet,
		RefreshTokenExpireTime: rtet,
	}, nil
}

// signing means
func (p *ProviderCrypto) SignAccessToken(payload model.JwtAccessTokenPayload) (string, error) {
	payload.ExpiresAt = jwt.NewNumericDate(time.Now().Add(p.AccessTokenExpireTime))
	payload.IssuedAt = jwt.NewNumericDate(time.Now())
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, payload)
	return token.SignedString(p.JwtAccessTokenEnckey)
}

func (p *ProviderCrypto) SignRefreshToken(payload model.JwtRefreshTokenPayload) (string, error) {
	// random jti: EdDSA is deterministic, so two tokens issued for the same
	// user within the same second would otherwise be byte-identical and
	// collide on the unique lookup_hash index
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	payload.ID = hex.EncodeToString(buf)
	payload.ExpiresAt = jwt.NewNumericDate(time.Now().Add(p.RefreshTokenExpireTime))
	payload.IssuedAt = jwt.NewNumericDate(time.Now())
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, payload)
	return token.SignedString(p.JwtRefreshTokenEnckey)
}

// mapTokenError converts raw jwt parse failures into business errors so the
// HTTP error handler can answer 401 with a readable message instead of
// falling back to a generic 500 (a *jwt.ValidationError carries no status).
func mapTokenError(err error) error {
	if errors.Is(err, jwt.ErrTokenExpired) {
		return errs.BuildErrUserLogin(
			errs.UserLoginTokenExpired,
			http.StatusUnauthorized,
			string(consts.ExprIndetermined),
		)
	}
	return errs.BuildErrUserLogin(
		errs.UserLoginTokenInvalid,
		http.StatusUnauthorized,
		string(consts.ExprIndetermined),
	)
}

func (p *ProviderCrypto) VerifyAccessToken(tokenStr string) (*model.JwtAccessTokenPayload, error) {
	var payload model.JwtAccessTokenPayload
	token, err := jwt.ParseWithClaims(tokenStr, &payload, NewJWTKeyFunc(p.JwtAccessTokenDeckey))
	if err != nil {
		return nil, mapTokenError(err)
	}
	if !token.Valid {
		return nil, mapTokenError(jwt.ErrTokenNotValidYet)
	}
	return &payload, nil
}

func (p *ProviderCrypto) VerifyRefreshToken(tokenStr string) (*model.JwtRefreshTokenPayload, error) {
	var payload model.JwtRefreshTokenPayload
	token, err := jwt.ParseWithClaims(tokenStr, &payload, NewJWTKeyFunc(p.JwtRefreshTokenDeckey))
	if err != nil {
		return nil, mapTokenError(err)
	}
	if !token.Valid {
		return nil, mapTokenError(jwt.ErrTokenNotValidYet)
	}
	return &payload, nil
}

func (p *ProviderCrypto) GetJwtAccessTokenDeckey() any {
	return p.JwtAccessTokenDeckey
}

func (p *ProviderCrypto) GetJwtRefreshTokenDeckey() any {
	return p.JwtRefreshTokenDeckey
}

func NewJWTKeyFunc(deckey ed25519.PublicKey) jwt.Keyfunc {
	return func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodEdDSA {
			return nil, errs.BuildErrValidation(
				errs.ValidationCryptoUnexpectedSigningMethod,
				http.StatusUnauthorized,
				string(consts.ExprDeckey),
				string(consts.ExprIndetermined),
			)
		}
		return deckey, nil
	}

}
