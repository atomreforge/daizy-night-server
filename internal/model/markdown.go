package model

import (
	"github.com/atomreforge/daizy-night-server/internal/consts"
	"gorm.io/gorm"
)

type Markdown struct {
	// mdid was cancelled and now type.Markdown is only a internal struct.
	//Mdid    uint `gorm:"uniqueIndex"`
	Title string `json:"md_title"`
	//Author  string `json:"md_author"`
	Content string `json:"md_content"`
	// cancelled
	//Type    consts.MarkdownItem
}

type Notification struct {
	gorm.Model
	Md           Markdown `gorm:"embedded;embeddedPrefix:md_"`
	Nid          uint     `gorm:"uniqueIndex"`
	TitlePreview string
	PostBy       uint
	PostAt       uint
	// for marking update ops, use `gorm.UpdateAt`
	Levels consts.Levels `gorm:"serializer:json"`
}
