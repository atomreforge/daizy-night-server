package user

import "time"

type ResponseAnyCalendarGet struct {
	UserID     uint                      `json:"uid"`
	CalendarID uint                      `json:"calendar_id"`
	Records    []ResponseAnyCalendarItem `json:"records"`
}

// dependently used.
type ResponseAnyCalendarItem struct {
	CalendarID uint         `json:"calendar_id"`
	Weekday    time.Weekday `json:"weekday"`
	StartMin   uint         `json:"start_min"`
	EndMin     uint         `json:"end_min"`
	Title      string       `json:"title"`
}
