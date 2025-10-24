package app

import (
	"gorm.io/gorm"
)

type Server struct {
	DB *gorm.DB
}

func NewServer(db *gorm.DB) *Server {
	return &Server{
		DB: db,
	}
}
