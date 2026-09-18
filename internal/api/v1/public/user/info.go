package user

import (
	"time"

	"github.com/atomreforge/daizy-night-server/internal/consts"
)

// ResponseAnyInfoGet renders the sanitized public view of a user's info:
// account-confidential fields (email, github identity) are deliberately
// omitted — same sanitization idea as the public calendar view dropping
// Roaming. No ownership check applies; any verified user may read it.
type ResponseAnyInfoGet struct {
	UserID       uint        `json:"uid"`
	Username     string      `json:"username"`
	Nickname     string      `json:"nickname"`
	RegisterTime time.Time   `json:"register_time"`
	Role         consts.Role `json:"role"`
}

