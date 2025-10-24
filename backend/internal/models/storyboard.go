package models

import (
	"time"
)

type ComicStoryboard struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	SectionID   uint      `json:"section_id" gorm:"not null"`
	ImagePrompt string    `json:"image_prompt" gorm:"type:text"`
	ImageURL    string    `json:"image_url" gorm:"size:500"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Section ComicSection              `json:"section,omitempty" gorm:"foreignKey:SectionID"`
	Details []ComicStoryboardDetail   `json:"details,omitempty" gorm:"foreignKey:StoryboardID"`
}

func (ComicStoryboard) TableName() string {
	return "comic_storyboard"
}

type ComicStoryboardDetail struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	StoryboardID uint      `json:"storyboard_id" gorm:"not null"`
	Detail       string    `json:"detail" gorm:"type:text"`
	RoleID       *uint     `json:"role_id"`
	TTSUrl       string    `json:"tts_url" gorm:"size:500"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Storyboard ComicStoryboard `json:"storyboard,omitempty" gorm:"foreignKey:StoryboardID"`
	Role       *ComicRole      `json:"role,omitempty" gorm:"foreignKey:RoleID"`
}

func (ComicStoryboardDetail) TableName() string {
	return "comic_storyboard_detail"
}
