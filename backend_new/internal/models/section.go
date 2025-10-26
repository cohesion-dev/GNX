package models

import "time"

type ComicSection struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	ComicID   uint      `gorm:"not null;index" json:"comic_id"`
	Title     string    `gorm:"" json:"title"`
	Index     int       `gorm:"not null" json:"index"`
	Content   string    `gorm:"type:text;not null" json:"-"`
	Status    string    `gorm:"default:'pending'" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Comic Comic       `gorm:"foreignKey:ComicID" json:"-"`
	Pages []ComicPage `gorm:"foreignKey:SectionID;orderBy:index" json:"pages,omitempty"`
}

func (ComicSection) TableName() string {
	return "comic_sections"
}
