package handler

import (
	"crypto/ed25519"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/atomreforge/daizy-night-server/internal/consts"
	"github.com/atomreforge/daizy-night-server/internal/errs"
	"github.com/atomreforge/daizy-night-server/internal/model"
	"github.com/atomreforge/daizy-night-server/internal/utils"
)

/* you mustn't trust frontend completely XD */

// it only check params and it's not its task to
func ValidateRegisterParams(b model.RegisterBody) error {
	// username
	if _, err := validateNonNull(b.Username); err != nil {
		return errs.BuildErrValidation(errs.ValidationKeyNull, http.StatusBadRequest, string(consts.ExprUsername), string(consts.ExprBlank))
	}

	if _, err := validateCharValid(b.Username); err != nil {
		return errs.BuildErrValidation(errs.ValidationKeyInvalidChar, http.StatusBadRequest, string(consts.ExprUsername), b.Username)
	}
	if _, err := validateLengthValid(b.Username, 1, 15); err != nil {
		return errs.BuildErrValidation(errs.ValidationKeyInvalidLength, http.StatusBadRequest, string(consts.ExprUsername), b.Username)
	}

	// nickname
	if _, err := validateNonNull(b.Nickname); err != nil {
		return errs.BuildErrValidation(errs.ValidationKeyNull, http.StatusBadRequest, string(consts.ExprNickname), string(consts.ExprBlank))
	}
	if _, err := validateLengthValid(b.Nickname, 1, 15); err != nil {
		return errs.BuildErrValidation(errs.ValidationKeyInvalidLength, http.StatusBadRequest, string(consts.ExprNickname), b.Nickname)
	}

	// password
	if _, err := validateNonNull(b.Password); err != nil {
		return errs.BuildErrValidation(errs.ValidationKeyNull, http.StatusBadRequest, string(consts.ExprPassword), string(consts.ExprBlank))
	}
	if _, err := validateLengthValid(b.Password, 6, 128); err != nil {
		return errs.BuildErrValidation(errs.ValidationKeyInvalidLength, http.StatusBadRequest, string(consts.ExprPassword), b.Password)
	}

	// registercode
	if ok, err := ValidateFormatRegistercode(b.Registercode); !ok {
		return err
	}

	/*// registercode
	if _, err := validateNonNull(b.Registercode); err != nil {
		return errs.BuildErrValidation(errs.ValidationKeyNull, http.StatusBadRequest, string(consts.ExprRegistercode), string(consts.ExprBlank))
	}*/

	return nil
}

func ValidateLoginParams(b *model.LoginBody) error {
	switch b.Loginway {
	case consts.LoginLegacy:
		//
		if _, err := validateNonNull(b.Username); err != nil {
			return errs.BuildErrValidation(errs.ValidationKeyNull, http.StatusBadRequest, string(consts.ExprUsername), string(consts.ExprBlank))
		}
		if _, err := validateNonNull(b.Password); err != nil {
			return errs.BuildErrValidation(errs.ValidationKeyNull, http.StatusBadRequest, string(consts.ExprPassword), string(consts.ExprBlank))
		}

	case consts.LoginGithub:
		return errs.BuildErrSupport(errs.FeatureUnsupported, http.StatusBadRequest)
		/* under constructing*/
	default:
		return errs.BuildErrValidation(errs.ValidationKeyNull, http.StatusBadRequest, string(consts.ExprLogin), string(b.Loginway))
	}

	return nil
}

func validateLengthValid(s string, min int, max int) (bool, error) {
	l := utf8.RuneCountInString(s)
	if l < min || l > max {
		return false, errs.BuildErrValidation(errs.ValidationKeyInvalidLength, http.StatusBadRequest, "", s)
	}
	return true, nil

}

func validateNonNull(v any) (bool, error) {
	if v == nil {
		return false, errs.BuildErrValidation(errs.ValidationKeyNull, http.StatusBadRequest, "", string(consts.ExprNull))
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String:
		if rv.String() == "" {
			return false, errs.BuildErrValidation(errs.ValidationKeyNull, http.StatusBadRequest, "", string(consts.ExprBlank))
		}
	case reflect.Slice, reflect.Map:
		if rv.Len() == 0 {
			return false, errs.BuildErrValidation(errs.ValidationKeyNull, http.StatusBadRequest, "", string(consts.ExprBlank))
		}
	}
	return true, nil
}

func validateCharValid(s string) (bool, error) {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !((c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9')) {
			return false, errs.BuildErrValidation(errs.ValidationKeyInvalidChar, http.StatusBadRequest, "", s)
		}
	}
	return true, nil
}

func ValidateFormatRegistercode(c model.RegistercodeRawHex) (bool, error) {
	if c == "" {
		return false, errs.BuildErrValidation(errs.ValidationKeyNull, http.StatusBadRequest, string(consts.ExprRegistercode), string(consts.ExprNull))
	}
	i := strings.LastIndex(string(c), ".")
	if i <= 0 || i == len(c)-1 {
		return false, errs.BuildErrValidation(errs.ValidationKeyBadFormat, http.StatusBadRequest, string(consts.ExprRegistercode), c.String())
	}
	payloadHex, sigHex := string(c)[:i], string(c)[i+1:]
	// an Ed25519 signature is exactly 64 bytes (128 hex chars); the payload
	// is bounded above. cheap size gates before any hex parsing or crypto.
	if len(sigHex) != 2*ed25519.SignatureSize ||
		len(payloadHex) == 0 || len(payloadHex) > maxRegistercodePayloadHex {
		return false, errs.BuildErrValidation(errs.ValidationKeyBadFormat, http.StatusBadRequest, string(consts.ExprRegistercode), c.String())
	}
	if !(utils.IsHex(payloadHex) && utils.IsHex(sigHex)) {
		return false, errs.BuildErrValidation(errs.ValidationKeyBadFormat, http.StatusBadRequest, string(consts.ExprRegistercode), c.String())
	}
	return true, nil
}

// ValidateSignoutParams checks the signout request. Until session identifiers
// are issued, an unspecified session defaults to the current session, which is
// only locatable through the presented refresh token — so it stays required,
// and a populated session list (signout-by-session) is rejected as unsupported.
func ValidateSignoutParams(b *model.SignoutBody) error {
	if len(b.Session) > 0 {
		return errs.BuildErrSupport(errs.FeatureUnsupported, http.StatusBadRequest)
	}

	if _, err := validateNonNull(b.RefreshToken); err != nil {
		return errs.BuildErrValidation(errs.ValidationKeyNull, http.StatusBadRequest, string(consts.JsonExprRefreshToken), string(consts.ExprBlank))
	}

	return nil
}

// ValidateRefreshParams checks the refresh request body before any service
// call: a missing token is a client validation problem (400), not a lookup
// miss (401).
func ValidateRefreshParams(refreshToken string) error {
	if _, err := validateNonNull(refreshToken); err != nil {
		return errs.BuildErrValidation(errs.ValidationKeyNull, http.StatusBadRequest, string(consts.JsonExprRefreshToken), string(consts.ExprBlank))
	}
	return nil
}

const (
	// hard cap on slots per timetable request, protects the db from abuse
	maxCalendarRecords = 200
	minutesPerDay      = 24 * 60

	// registercode hex bounds: the payload is a tiny JSON (magicword +
	// expiry; real codes are ~160 hex chars), the signature is fixed-size
	maxRegistercodePayloadHex = 2048
)

// ValidateCalendarPutParams checks a full timetable replacement. an empty
// records list is valid (it clears the timetable). every slot must stay
// within one day: 0 <= start_min < end_min <= 1440.
func ValidateCalendarPutParams(b *model.CalendarPutBody) error {
	if len(b.Records) > maxCalendarRecords {
		return errs.BuildErrValidation(errs.ValidationKeyInvalidLength, http.StatusBadRequest, string(consts.JsonExprRecords), fmt.Sprintf("%d", len(b.Records)))
	}
	for _, it := range b.Records {
		if it.Weekday < time.Sunday || it.Weekday > time.Saturday {
			return errs.BuildErrValidation(errs.ValidationKeyBadFormat, http.StatusBadRequest, string(consts.JsonExprWeekday), fmt.Sprintf("%d", it.Weekday))
		}
		// start_min non-negativity is enforced by its uint type at bind
		// time (negative literals fail JSON unmarshalling)
		if it.EndMin > minutesPerDay || it.StartMin >= it.EndMin {
			return errs.BuildErrValidation(errs.ValidationKeyBadFormat, http.StatusBadRequest, string(consts.JsonExprStartMin)+"::"+string(consts.JsonExprEndMin), fmt.Sprintf("%d..%d", it.StartMin, it.EndMin))
		}
		if _, err := validateNonNull(it.Title); err != nil {
			return errs.BuildErrValidation(errs.ValidationKeyNull, http.StatusBadRequest, string(consts.JsonExprTitle), string(consts.ExprBlank))
		}
		if _, err := validateLengthValid(it.Title, 1, 255); err != nil {
			return errs.BuildErrValidation(errs.ValidationKeyInvalidLength, http.StatusBadRequest, string(consts.JsonExprTitle), it.Title)
		}
	}
	return nil
}
