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
	Sudo()
}

type InterfaceServiceHealth interface {
	HealthCheckDb() (bool, error)
}
