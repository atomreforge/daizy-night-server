package model

import (
	"github.com/atomreforge/daizy-night-server/internal/consts"
	"gorm.io/gorm"
)

type Markdown struct {
	gorm.Model
	Title   string
	Content string
}

type Notification struct {
	Md           Markdown
	TitlePreview string
	Author       string
	Levels       consts.Levels
}
