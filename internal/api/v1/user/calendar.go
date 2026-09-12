package v1

import (
	"time"

	"github.com/atomreforge/daizy-night-server/internal/model"
)

// ResponseUserCalendarGet renders a timetable without leaking internal ids or
// soft-delete bookkeeping (explicit mapping, same idea as InfoMe).

type ResponseUserCalendarGet struct {
	Roaming    model.Roaming              `json:"roaming"`
	UserID     uint                       `json:"uid"`
	CalendarID uint                       `json:"calendar_id"`
	Records    []ResponseUserCalendarItem `json:"records"`
}

// dependently used.
type ResponseUserCalendarItem struct {
	Roaming    model.Roaming `json:"roaming"`
	CalendarID uint          `json:"calendar_id"`
	Weekday    time.Weekday  `json:"weekday"`
	StartMin   uint          `json:"start_min"`
	EndMin     uint          `json:"end_min"`
	Title      string        `json:"title"`
}

type ResponseUserCalendarPut struct {
	Message string `json:"message"`
}

type ResponseUserCalendarDelete struct {
	Message string `json:"message"`
}
