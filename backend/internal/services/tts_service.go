package services

import (
	"fmt"

	"github.com/cohesion-dev/GNX/backend/internal/repositories"
)

type TTSService struct {
	storyboardRepo *repositories.StoryboardRepository
}

func NewTTSService(storyboardRepo *repositories.StoryboardRepository) *TTSService {
	return &TTSService{
		storyboardRepo: storyboardRepo,
	}
}

func (s *TTSService) GetTTSAudio(detailID uint) (string, error) {
	detail, err := s.storyboardRepo.GetDetailByID(detailID)
	if err != nil {
		return "", fmt.Errorf("TTS detail not found: %w", err)
	}

	if detail.TTSUrl == "" {
		return "", fmt.Errorf("TTS audio not yet generated")
	}

	return detail.TTSUrl, nil
}
