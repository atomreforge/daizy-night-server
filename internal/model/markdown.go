package model

import (
	"github.com/atomreforge/daizy-night-server/internal/consts"
	"gorm.io/gorm"
)

type Markdown struct {
	// mdid was cancelled and now type.Markdown is only a internal struct.
	//Mdid    uint `gorm:"uniqueIndex"`
	Title   string
	Content string
	Type    consts.MarkdownItem
}

type Notification struct {
	gorm.Model
	Md           Markdown
	Nid          uint `gorm:"uniqueIndex"`
	TitlePreview string
	Author       string
	Levels       consts.Levels
}
