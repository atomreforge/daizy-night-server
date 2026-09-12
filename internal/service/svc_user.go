package service

import (
	//"errors"
	"fmt"
	"log/slog"
	"net/http"

	abstract "github.com/atomreforge/daizy-night-server/internal/abstract/interface"
	"github.com/atomreforge/daizy-night-server/internal/consts"

	//"github.com/atomreforge/daizy-night-server/internal/crypto"
	"github.com/atomreforge/daizy-night-server/internal/errs"
	"github.com/atomreforge/daizy-night-server/internal/model"
	"github.com/atomreforge/daizy-night-server/internal/utils"
	//"gorm.io/gorm"
)

var _ abstract.InterfaceServiceUser = (*ServiceUser)(nil)

// shall use Service to avoid direct operation on Repo if there is already a service to provide the needed functionality.
type ServiceUser struct {
	repoUser     abstract.InterfaceRepoUser
	repoToken    abstract.InterfaceRepoToken
	repoCalendar abstract.InterfaceRepoCalendar
	//repoRegcode abstract.InterfaceRepoRegistercode
	pSvcCode abstract.InterfaceServiceCode
	crypto   abstract.InterfaceCrypto
}

func NewServiceUser(
	repoUser abstract.InterfaceRepoUser,
	repoToken abstract.InterfaceRepoToken,
	repoCalendar abstract.InterfaceRepoCalendar,
	//repoRegcode abstract.InterfaceRepoRegistercode,
	pSvcCode abstract.InterfaceServiceCode,
	crypto abstract.InterfaceCrypto,
) *ServiceUser {
	return &ServiceUser{
		repoUser:     repoUser,
		repoToken:    repoToken,
		repoCalendar: repoCalendar,
		//repoRegcode: repoRegcode,
		pSvcCode: pSvcCode,
		crypto:   crypto,
	}
}

func (s *ServiceUser) Register(b *model.RegisterBody) (*model.User, error) {
	slog.Info("service processing register ...")

	// router
	switch b.Registerway {
	case consts.RegisterLegacy:
		payload, err := s.crypto.AnalyzeRegistercode(b.Registercode)
		if err != nil {
			return nil, err
		} //it can only be ErrRegistercode

		// verify before hashing
		if !s.crypto.VerifyRegistercodePayload(*payload) {
			return nil, errs.BuildErrRegistercode(errs.RegistercodeUnusableOutdated, http.StatusBadRequest)
		}

		// audit: record only registercodes whose signature and expiry both
		// check out. recording attacker-controlled garbage before validation
		// turned the audit table into an unauthenticated storage-filling vector.
		if err := s.pSvcCode.RecordNewRegistercode(&model.RegistercodeRecord{
			RawHex: b.Registercode,
			Used:   false,
		}); err != nil {
			return nil, err
		}

		// salt psw after hashing; hashing is perf-expensive
		passwordHash, err := utils.HashCreate(b.Password)
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}

		err = s.repoUser.CreateUser(&model.User{
			Username:     b.Username,
			Nickname:     b.Nickname,
			PasswordHash: passwordHash,
			Registercode: b.Registercode,
		})

		if err != nil {
			return nil, err
		}

	case consts.RegisterGithub:
		return nil, &errs.ErrSupport{
			Type: errs.FeatureUnsupported,
			Http: http.StatusInternalServerError,
		}
	default:
		slog.Error(string(consts.ExprUnreachableCase))
		return nil, &errs.ErrUnknown{
			Type: errs.Undefined,
			Http: http.StatusInternalServerError,
		}
	}

	// DO NOT REMOVE
	user, err := s.repoUser.GetUserByUsername(b.Username)
	if err != nil {
		slog.Error(string(consts.ExprUnreachableCase))
		return nil, err
	}
	return user, nil
}

func (s *ServiceUser) Login(b *model.LoginBody) (success bool, accessToken string, refreshToken string, err error) {
	switch b.Loginway {
	case consts.LoginLegacy:
		{
			if b.Username != "" {
				user, err := s.repoUser.GetUserByUsername(b.Username)
				// db error during GetUser
				if err != nil {
					errapp, ok := errs.Easx[*errs.ErrDbRecord](err)
					if ok {
						if errapp.Type == errs.DbRecordUsernameNotFound {
							return false, "", "", errs.BuildErrUserLogin(errs.UserLoginParamsIncorrect, http.StatusBadRequest, string(consts.ExprIndetermined))
						}
					}
					return false, "", "", err
				}

				matched, err := utils.HashVerify(b.Password, user.PasswordHash)
				if err != nil {
					return false, "", "", err
				}
				if !matched {
					return false, "", "", errs.BuildErrUserLogin(errs.UserLoginParamsIncorrect, http.StatusBadRequest, user.Username)
				}

				payloadAccessToken := model.JwtAccessTokenPayload{
					Uid:      user.UserID,
					Username: user.Username,
					Role:     user.Role,
				}
				payloadRefreshToken := model.JwtRefreshTokenPayload{
					Uid:      user.UserID,
					Username: user.Username,
					Role:     user.Role,
				}

				accessToken, err := s.crypto.SignAccessToken(payloadAccessToken)
				if err != nil {
					slog.Error(err.Error())
					return false, "", "", err
				}

				refreshToken, err := s.crypto.SignRefreshToken(payloadRefreshToken)
				if err != nil {
					slog.Error(err.Error())
					return false, "", "", err
				}

				if err := s.repoToken.SaveRefreshToken(user.UserID, refreshToken); err != nil {
					slog.Error(err.Error())
					return false, "", "", err
				}

				return true, accessToken, refreshToken, nil

			}
			return false, "", "", errs.BuildErrValidation(errs.ValidationKeyNull, http.StatusBadRequest, string(consts.ExprUsername), string(consts.ExprNull))
		}
	// under constructing ..
	case consts.LoginGithub:
		{
			return false, "", "", errs.BuildErrSupport(errs.FeatureUnsupported, http.StatusBadRequest)
		}
	}

	// wont arrive here. (hope so?)
	return false, "", "", errs.BuildErrSupport(errs.FeatureUnsupported, http.StatusBadRequest)
}

/*
//repalced

	func (s *ServiceUser) RefreshAccessToken(rawToken string) (success bool, accessToken string, refreshToken string, err error) {
		payload, err := s.crypto.VerifyRefreshToken(rawToken)
		if err != nil {
			return false, "", "", err
		}

		valid, err := s.repoToken.GetRefreshToken(payload.Uid, rawToken)
		if err != nil {
			return false, "", "", err
		}

		if !valid {
			return false, "", "", errs.BuildErrUserLogin(errs.UserLoginParamsIncorrect, http.StatusBadRequest, string(consts.ExprIndetermined))
		}

		if err = s.repoToken.RevokeUserTokens(payload.Uid); err != nil {
			return false, "", "", err
		}

		payloadAccess := model.JwtAccessTokenPayload{
			Uid:      payload.Uid,
			Username: payload.Username,
			Role:     payload.Role,
		}
		accessToken, err = s.crypto.SignAccessToken(payloadAccess)
		if err != nil {
			return false, "", "", err
		}

		payloadRefresh := model.JwtRefreshTokenPayload{
			Uid:      payload.Uid,
			Username: payload.Username,
			Role:     payload.Role,
		}
		refreshToken, err = s.crypto.SignRefreshToken(payloadRefresh)
		if err != nil {
			return false, "", "", err
		}

		err = s.repoToken.SaveRefreshToken(payload.Uid, refreshToken)
		if err != nil {
			return false, "", "", err
		}
		return true, accessToken, refreshToken, nil
	}
*/
func (s *ServiceUser) RefreshAccessToken(rawToken string) (success bool, accessToken string, refreshToken string, err error) {
	inputPayload, err := s.crypto.VerifyRefreshToken(rawToken)
	if err != nil {
		return false, "", "", err
	}

	row, err := s.repoToken.GetRefreshToken(inputPayload.Uid, rawToken)
	if err != nil {
		return false, "", "", err
	}
	if row == nil {
		return false, "", "", errs.BuildErrUserLogin(
			errs.UserLoginParamsIncorrect,
			http.StatusUnauthorized,
			string(consts.ExprIndetermined),
		)
	}

	// pure computations before any DB write: a signing failure leaves
	// the database untouched and the old token valid
	payloadAccess := model.JwtAccessTokenPayload{
		Uid:      inputPayload.Uid,
		Username: inputPayload.Username,
		Role:     inputPayload.Role,
	}
	accessToken, err = s.crypto.SignAccessToken(payloadAccess)
	if err != nil {
		return false, "", "", err
	}

	payloadRefresh := model.JwtRefreshTokenPayload{
		Uid:      inputPayload.Uid,
		Username: inputPayload.Username,
		Role:     inputPayload.Role,
	}
	refreshToken, err = s.crypto.SignRefreshToken(payloadRefresh)
	if err != nil {
		return false, "", "", err
	}

	// atomic rotation: revoke the used token + save the new one
	if err = s.repoToken.RotateRefreshToken(inputPayload.Uid, row.LookupHash, refreshToken); err != nil {
		return false, "", "", err
	}
	defer slog.Info(fmt.Sprintf("successfully rotated/refreshed refresh-token; user::%s; uid::%d", inputPayload.Username, inputPayload.Uid))
	return true, accessToken, refreshToken, nil
}

// Signout revokes the refresh token presented by the user — the credential
// that identifies the current session until real session identifiers exist.
// The uid arrives pre-filled by the handler from the JWT claims. Idempotent:
// unknown, invalid or already-revoked tokens resolve to a successful no-op
// so signout never fails from the client's perspective.
func (s *ServiceUser) Signout(b *model.SignoutBody) (success bool, err error) {
	if _, err = s.crypto.VerifyRefreshToken(b.RefreshToken); err != nil {
		return true, nil
	}
	if err = s.repoToken.RevokeRefreshToken(b.Uid, b.RefreshToken); err != nil {
		return false, err
	}
	slog.Info("successfully signed out", "uid", b.Uid)
	return true, nil
}

func (s *ServiceUser) GetInfoMineByUid(uid uint) (*model.InfoUser, error) {
	// the repo answers 404 for a missing user record (see repo contract)
	user, err := s.repoUser.GetUserByUid(uid)
	if err != nil {
		return nil, err
	}

	return &model.InfoUser{
		Uid:          user.UserID,
		Username:     user.Username,
		Nickname:     user.Nickname,
		Email:        user.Email,
		RegisterTime: user.RegisterTime,
		Role:         user.Role,
		GithubID:     user.GithubID,
		GithubLogin:  user.GithubLogin,
	}, nil
}

func (s *ServiceUser) GetUserByUsername(name string) (*model.User, error) {
	// the repo answers 400 for a missing username (see repo contract)
	user, err := s.repoUser.GetUserByUsername(name)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *ServiceUser) GetUserByUid(uid uint) (*model.User, error) {
	// the repo answers 404 for a missing user record (see repo contract)
	user, err := s.repoUser.GetUserByUid(uid)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *ServiceUser) GetUidByRefreshToken(rawToken string) (uint, error) {
	return s.repoToken.GetUidByRefreshToken(rawToken)
}

func (s *ServiceUser) GetUserByRefreshToken(rawToken string) (*model.User, error) {
	if _, err := s.crypto.VerifyRefreshToken(rawToken); err != nil {
		return nil, err
	}

	uid, err := s.repoToken.GetUidByRefreshToken(rawToken)
	if err != nil {
		return nil, err
	}

	return s.repoUser.GetUserByUid(uid)
}

func (s *ServiceUser) AddCalendar(cal *model.CalendarTable) error {
	return s.repoCalendar.RecordNewCalendar(cal)
}

func (s *ServiceUser) UpdateCalendar(cal *model.CalendarTable) error {
	return s.repoCalendar.UpdateCalendar(cal)
}

func (s *ServiceUser) RemoveCalendarByModel(cal *model.CalendarTable) error {
	return s.repoCalendar.RemoveCalendarByModel(cal)
}

func (s *ServiceUser) RemoveCalendarByUid(uid uint) error {
	return s.repoCalendar.RemoveCalendarByUid(uid)
}

func (s *ServiceUser) GetCalendarByUid(uid uint) (*model.CalendarTable, error) {
	return s.repoCalendar.GetCalendarByUid(uid)
}

func (s *ServiceUser) GetCalendarByUsername(name string) (*model.CalendarTable, error) {
	user, err := s.repoUser.GetUserByUsername(name)
	if err != nil {
		if e, ok := errs.Easx[*errs.ErrDbRecord](err); ok && e.Type == errs.DbRecordUsernameNotFound {
			return nil, errs.BuildErrDbRecord(errs.DbRecordUsernameNotFound, http.StatusNotFound, string(consts.InlineExprUser))
		}
		return nil, err
	}
	return s.repoCalendar.GetCalendarByUid(user.UserID)
}
