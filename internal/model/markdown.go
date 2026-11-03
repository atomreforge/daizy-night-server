package model

import (
	"github.com/atomreforge/daizy-night-server/internal/consts"
	"gorm.io/gorm"
)

type Markdown struct {
	gorm.Model
	Mdid    uint `gorm:"uniqueIndex"`
	Title   string
	Content string
	Type    consts.MarkdownItem
}

type Notification struct {
	Md           Markdown
	TitlePreview string
	Author       string
	Levels       consts.Levels
}
