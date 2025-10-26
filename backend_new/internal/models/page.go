package models

import "time"

type ComicPage struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	SectionID   uint      `gorm:"not null;index" json:"section_id"`
	Index       int       `gorm:"not null" json:"index"`
	ImagePrompt string    `gorm:"type:text" json:"-"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Section ComicSection      `gorm:"foreignKey:SectionID" json:"-"`
	Details []ComicPageDetail `gorm:"foreignKey:PageID;orderBy:index" json:"details,omitempty"`
}

func (ComicPage) TableName() string {
	return "comic_pages"
}
