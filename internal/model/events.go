package model

import "github.com/atomreforge/daizy-night-server/internal/consts"

type Event struct {
	Title     string
	Name      string // if any
	Level     consts.Levels
	EventType consts.EventType
}
