package model

import (
	"time"

	"github.com/atomreforge/daizy-night-server/internal/consts"
)

type RegisterBody struct {
	Registerway  consts.Registerway `json:"registerway"`
	Username     string             `json:"username"`
	Nickname     string             `json:"nickname"`
	Password     string             `json:"password"`
	Registercode RegistercodeRawHex `json:"registercode"`
}

type LoginBody struct {
	Loginway  consts.Loginway `json:"loginway"`
	Username  string          `json:"username"`
	Password  string          `json:"password"`
	Entrycode string          `json:"entrycode"`
}

type InfoUser struct {
	Uid uint `json:"uid"`

	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`

	RegisterTime time.Time `json:"register_time"`

	Role consts.Role `json:"role"`

	GithubID    *int64  `json:"github_id"`
	GithubLogin *string `json:"github_login"`
}

// SignoutBody carries the current session's signout request. Uid is NOT part
// of the request payload: the handler fills it from the authenticated JWT
// claims, so a client can never sign out on behalf of someone else.
type SignoutBody struct {
	Uid uint `json:"-"`
	// the credential that identifies the current session until real
	// session identifiers exist
	RefreshToken string `json:"refresh_token"`
	// placeholder for future session identifiers: an unspecified session
	// defaults to the current session, so it must stay empty or omitted
	// until signout-by-session is actually implemented
	Session []string `json:"session"`
}

// CalendarItemBody is one weekly slot of a calendar-upsert request.
type CalendarItemBody struct {
	Weekday  time.Weekday `json:"weekday"`
	StartMin uint         `json:"start_min"`
	EndMin   uint         `json:"end_min"`
	Title    string       `json:"title"`
}

// CalendarPutBody carries a full timetable replacement. UserID is NOT part
// of the request payload: the handler fills it from the authenticated JWT
// claims (like SignoutBody.Uid), and the :username path segment is checked
// against the identity as well.
type CalendarPutBody struct {
	UserID  uint               `json:"-"`
	Records []CalendarItemBody `json:"records"`
}
