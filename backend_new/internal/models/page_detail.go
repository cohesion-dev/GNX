package models

import "time"

type ComicPageDetail struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	PageID    uint      `gorm:"not null;index" json:"page_id"`
	Index     int       `gorm:"not null" json:"index"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	RoleID    *uint     `gorm:"index" json:"role_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Page ComicPage  `gorm:"foreignKey:PageID" json:"-"`
	Role *ComicRole `gorm:"foreignKey:RoleID" json:"-"`
}

func (ComicPageDetail) TableName() string {
	return "comic_page_details"
}
