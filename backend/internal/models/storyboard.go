package models

import (
	"time"
)

type ComicStoryboard struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	SectionID   uint      `json:"section_id" gorm:"not null"`
	Index       int       `json:"index" gorm:"not null"`
	ImagePrompt string    `json:"image_prompt" gorm:"type:text"`
	ImageURL    string    `json:"image_url" gorm:"size:500"`
	Status      string    `json:"status" gorm:"size:20;default:'pending'"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Section ComicSection            `json:"section,omitempty" gorm:"foreignKey:SectionID"`
	Details []ComicStoryboardDetail `json:"details,omitempty" gorm:"foreignKey:StoryboardID"`
}

func (ComicStoryboard) TableName() string {
	return "comic_storyboard"
}

type ComicStoryboardDetail struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	StoryboardID uint      `json:"storyboard_id" gorm:"not null"`
	Index        int       `json:"index" gorm:"not null"`
	Text         string    `json:"text" gorm:"type:text"`
	VoiceName    string    `json:"voice_name" gorm:"size:100"`
	VoiceType    string    `json:"voice_type" gorm:"size:100"`
	SpeedRatio   float64   `json:"speed_ratio" gorm:"default:1.0"`
	IsNarration  bool      `json:"is_narration" gorm:"default:false"`
	TTSUrl       string    `json:"tts_url" gorm:"size:500"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Storyboard ComicStoryboard `json:"storyboard,omitempty" gorm:"foreignKey:StoryboardID"`
}

func (ComicStoryboardDetail) TableName() string {
	return "comic_storyboard_detail"
}
