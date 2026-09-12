package model

import (
	"time"

	"gorm.io/gorm"
)

type CalendarTable struct {
	gorm.Model
	Roaming
	UserID uint `gorm:"not null;uniqueIndex"` // single calendar per user
	// calendarID is the unique identifier for the calendar. length 7 digits.
	CalendarID uint           `gorm:"not null;unique"`
	Records    []CalendarItem `gorm:"foreignKey:CalendarID;references:CalendarID;constraint:OnDelete:CASCADE"`
}

type Roaming struct {
	Description string `json:"description"`
	Annotation  string `json:"annotation"`
}

// a lesson or event in a calendar: weekday + time range in minutes-of-day
// (e.g. 480 == 08:00), plus a title. the service layer must ensure
// 0 <= StartMin < EndMin <= 24*60.
type CalendarItem struct {
	gorm.Model
	Roaming
	CalendarID uint         `gorm:"not null;index:idx_calendar_weekday,priority:1"`
	Weekday    time.Weekday `gorm:"not null;index:idx_calendar_weekday,priority:2"`
	// minute count in a day
	StartMin uint   `gorm:"not null"`
	EndMin   uint   `gorm:"not null"`
	Title    string `gorm:"not null;size:255"`
}
