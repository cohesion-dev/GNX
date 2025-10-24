package models

import (
	"time"
)

type Storyboard struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	SectionID   uint      `gorm:"not null;index" json:"section_id"`
	ImagePrompt string    `gorm:"type:text" json:"image_prompt"`
	ImageURL    string    `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	Details []StoryboardDetail `gorm:"foreignKey:StoryboardID" json:"details,omitempty"`
}

type StoryboardDetail struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	StoryboardID uint      `gorm:"not null;index" json:"storyboard_id"`
	Detail       string    `gorm:"type:text" json:"detail"`
	RoleID       *uint     `gorm:"index" json:"role_id"`
	TTSURL       string    `json:"tts_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type RoleStoryboard struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	StoryboardID uint      `gorm:"not null;index" json:"storyboard_id"`
	RoleID       uint      `gorm:"not null;index" json:"role_id"`
	CreatedAt    time.Time `json:"created_at"`
}
