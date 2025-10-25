package services

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cohesion-dev/GNX/backend/internal/repositories"
	"github.com/cohesion-dev/GNX/backend/pkg/storage"
)

type TTSService struct {
	storyboardRepo *repositories.StoryboardRepository
	storageClient  *storage.QiniuClient
}

func NewTTSService(storyboardRepo *repositories.StoryboardRepository, storageClient *storage.QiniuClient) *TTSService {
	return &TTSService{
		storyboardRepo: storyboardRepo,
		storageClient:  storageClient,
	}
}

func (s *TTSService) GetTTSAudio(detailID uint) ([]byte, error) {
	detail, err := s.storyboardRepo.GetDetailByID(detailID)
	if err != nil {
		return nil, fmt.Errorf("TTS detail not found: %w", err)
	}

	if detail.TTSUrl != "" {
		return s.fetchAudioFromURL(detail.TTSUrl)
	}

	voice := "default"
	if detail.Role != nil && detail.Role.Voice != "" {
		voice = detail.Role.Voice
	}

	audioData, err := s.storageClient.GenerateTTS(detail.Detail, voice)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TTS: %w", err)
	}

	go s.saveAudioToQiniu(detailID, audioData)

	return audioData, nil
}

func (s *TTSService) fetchAudioFromURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch audio from URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch audio, status: %d", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}

func (s *TTSService) saveAudioToQiniu(detailID uint, audioData []byte) {
	key := fmt.Sprintf("tts/%d.mp3", detailID)
	audioURL, err := s.storageClient.UploadAudio(key, audioData)
	if err != nil {
		return
	}

	s.storyboardRepo.UpdateDetailTTSURL(detailID, audioURL)
}
