package abstract

import (
	"github.com/atomreforge/daizy-night-server/internal/model"
)

type InterfaceRepoUser interface {
	CreateUser(b *model.User) error
	GetUserByUsername(name string) (*model.User, error)
	GetUserByUid(uid uint) (*model.User, error)
}

type InterfaceRepoToken interface {
	SaveRefreshToken(uid uint, rawToken string) error
	GetRefreshToken(uid uint, rawToken string) (*model.RefreshToken, error)
	// notice that this func will revoke all tokens of a user
	RevokeUserTokens(uid uint) error
	// revokes a single token; unknown/already-revoked tokens affect zero rows (idempotent)
	RevokeRefreshToken(uid uint, rawToken string) error
	RotateRefreshToken(uid uint, usedLookupHash, newRawToken string) error
	GetUidByRefreshToken(rawToken string) (uint, error)
}

type InterfaceRepoRegistercode interface {
	//Record(registercodeRaw model.RegistercodeRawHex) error
	RecordNewRegistercode(b *model.RegistercodeRecord) error
	Remove(registercodeRaw model.RegistercodeRawHex) error
	Used(registercodRawHex model.RegistercodeRawHex, value bool, userID uint) error
	Updates(record model.RegistercodeRecord) error
	GetRecordByRegistercode(rawHex model.RegistercodeRawHex) (*model.RegistercodeRecord, error)
}

type InterfaceRepoCalendar interface {
	// insert a new calendar (with items) for a user; generates CalendarID
	RecordNewCalendar(b *model.CalendarTable) error
	// fetch the user's calendar with all items preloaded
	GetCalendarByUid(uid uint) (*model.CalendarTable, error)
	// full replacement of the user's items, auto lookup by uid; the row is
	// created on first use
	UpdateCalendar(b *model.CalendarTable) error
	// remove by uid, cascading the items; idempotent
	RemoveCalendarByUid(uid uint) error
	RemoveCalendarByModel(b *model.CalendarTable) error
}

type InterfaceRepoNotification interface {
	// inserts a notification; `Nid` is always server-allocated as a strictly
	// positive, monotonically increasing integer (never zero, never reused)
	CreateItem(n model.Notification) error
	RemoveItem(nid uint) error
	GetItemByNid(nid uint) (model.Notification, error)
	UpdateItem(n model.Notification) error
	// greatest live nid; 0 when the table holds no live rows
	GetLatestNotificationNid() (uint, error)
	// closed range [nidBegin, nidEnd], ascending by nid; never a nil slice
	GetNotificationBatch(nidBegin, nidEnd uint) (*[]model.Notification, error)
	// stamp post_by/post_at (publish); unknown or removed nid -> 404-class error
	PostItem(nid uint, postBy uint) error
	// clear post_by/post_at (unpublish); unknown or removed nid -> 404-class error
	WithdrawItem(nid uint) error
}
