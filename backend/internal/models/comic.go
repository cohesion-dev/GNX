package models

import (
	"time"
)

type Comic struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	Title          string     `json:"title" gorm:"size:255;not null;uniqueIndex"`
	Brief          string     `json:"brief" gorm:"type:text"`
	Icon           string     `json:"icon" gorm:"size:500"`
	IconPrompt     string     `json:"icon_prompt" gorm:"type:text"`
	Bg             string     `json:"bg" gorm:"size:500"`
	BgPrompt       string     `json:"bg_prompt" gorm:"type:text"`
	NovelFileURL   string     `json:"novel_file_url" gorm:"size:500"`
	UserPrompt     string     `json:"user_prompt" gorm:"type:text"`
	Status         string     `json:"status" gorm:"size:20;default:'pending'"`
	HasMoreContent bool       `json:"has_more_content" gorm:"default:true"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" gorm:"index"`

	Roles    []ComicRole    `json:"roles,omitempty" gorm:"foreignKey:ComicID"`
	Sections []ComicSection `json:"sections,omitempty" gorm:"foreignKey:ComicID"`
}

func (Comic) TableName() string {
	return "comic"
}
