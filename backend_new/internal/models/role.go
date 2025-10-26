package models

import "time"

type ComicRole struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	ComicID   uint      `gorm:"not null;index" json:"comic_id"`
	Name      string    `gorm:"not null" json:"name"`
	Brief     string    `gorm:"type:text" json:"brief"`
	ImageID   string    `gorm:"" json:"image_id"`
	Gender    string    `gorm:"" json:"gender"`
	Age       string    `gorm:"" json:"age"`
	VoiceName string    `gorm:"" json:"voice_name"`
	VoiceType string    `gorm:"" json:"voice_type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Comic Comic `gorm:"foreignKey:ComicID" json:"-"`
}

func (ComicRole) TableName() string {
	return "comic_roles"
}
