package abstract

import "github.com/atomreforge/daizy-night-server/internal/model"

type InterfaceServiceUser interface {
	Register(b *model.RegisterBody) (*model.User, error)
	Login(b *model.LoginBody) (success bool, accessToken string, refreshToken string, err error)
	Signout(b *model.SignoutBody) (success bool, err error)
	RefreshAccessToken(rawToken string) (success bool, accessToken string, refreshToken string, err error)

	GetUserByUid(uid uint) (*model.User, error)
	GetUserByUsername(name string) (*model.User, error)
	GetInfoMineByUid(uid uint) (*model.InfoUser, error)
	GetInfoByUsername(name string) (*model.InfoUser, error)

	GetUidByRefreshToken(rawToken string) (uint, error)

	AddCalendar(cal *model.CalendarTable) error
	UpdateCalendar(cal *model.CalendarTable) error
	RemoveCalendarByModel(cal *model.CalendarTable) error
	RemoveCalendarByUid(uid uint) error
	GetCalendarByUid(uid uint) (*model.CalendarTable, error)
	GetCalendarByUsername(name string) (*model.CalendarTable, error)
}

type InterfaceServiceCode interface {
	RecordNewRegistercode(registercodeRecord *model.RegistercodeRecord) error
	RemoveRegistercode(registercodeRaw model.RegistercodeRawHex) error
}

type InterfaceServiceAdmin interface {
	// test method
	Sudo()
}

type InterfaceServiceHealth interface {
	// check whether db is ok or not
	HealthCheckDb() (bool, error)
}

type InterfaceServiceNotification interface {
	CreateNotification(n model.Notification) error

	// not physically delete from repo
	DeleteNotification(nid uint) error

	UpdateNotification(n model.Notification) error

	// withdraw a posted notification
	WithdrawNotification(nid uint) error

	// backend compares the latest nid in repo with that client has; execute sync ops.
	GetNotificationBatchUpdate(clientLatestNid uint) (*[]model.Notification, error)

	// get notif with specific `nid`
	GetNotificationSingleByNid(nid uint) (*model.Notification, error)

	// get notif with specific range of `nid`
	GetNotificationBatchByNid(nidBegin uint, nidEnd uint) (*[]model.Notification, error)

	// get the latest `nid` from repo
	GetLatestNotificationNid() (uint, error)

	// post a notif to clients after created.
	PostNotification(n *model.Notification) error

	// upload and post a single notif
	//UploadAndPostNotification(n *model.Notification) error
}
