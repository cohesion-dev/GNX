package models

import (
	"time"
)

type ComicRole struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	ComicID      uint      `json:"comic_id" gorm:"not null"`
	Name         string    `json:"name" gorm:"size:100;not null"`
	Gender       string    `json:"gender" gorm:"size:20"`
	Age          string    `json:"age" gorm:"size:50"`
	ImageURL     string    `json:"image_url" gorm:"size:500"`
	Brief        string    `json:"brief" gorm:"type:text"`
	Hair         string    `json:"hair" gorm:"type:text"`
	HabitualExpr string    `json:"habitual_expr" gorm:"type:text"`
	SkinTone     string    `json:"skin_tone" gorm:"type:text"`
	FaceShape    string    `json:"face_shape" gorm:"type:text"`
	VoiceName    string    `json:"voice_name" gorm:"size:100"`
	VoiceType    string    `json:"voice_type" gorm:"size:100"`
	SpeedRatio   float64   `json:"speed_ratio" gorm:"default:1.0"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Comic Comic `json:"comic,omitempty" gorm:"foreignKey:ComicID"`
}

func (ComicRole) TableName() string {
	return "comic_role"
}
