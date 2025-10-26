package services

import (
	"fmt"
	"time"

	"github.com/cohesion-dev/GNX/backend_new/pkg/storage"
)

type ImageService struct {
	storage *storage.Storage
}

func NewImageService(storage *storage.Storage) *ImageService {
	return &ImageService{storage: storage}
}

func (s *ImageService) GetImageURL(imageID string) (string, error) {
	if imageID == "" {
		return "", fmt.Errorf("image not found")
	}

	url := s.storage.GetPrivateURL(imageID, time.Hour)
	return url, nil
}
