package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cohesion-dev/GNX/backend_new/config"
	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/storage"
)

type Storage struct {
	cfg    *config.StorageConfig
	mac    *auth.Credentials
	bucket string
	domain string
}

func NewStorage(cfg *config.StorageConfig) *Storage {
	mac := auth.New(cfg.AccessKey, cfg.SecretKey)
	return &Storage{
		cfg:    cfg,
		mac:    mac,
		bucket: cfg.Bucket,
		domain: cfg.Domain,
	}
}

func (s *Storage) UploadBytes(data []byte, key string) error {
	putPolicy := storage.PutPolicy{
		Scope: fmt.Sprintf("%s:%s", s.bucket, key),
	}
	upToken := putPolicy.UploadToken(s.mac)

	cfg := storage.Config{
		UseCdnDomains: false,
		UseHTTPS:      true,
	}

	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}

	dataLen := int64(len(data))
	err := formUploader.Put(context.Background(), &ret, upToken, key, bytes.NewReader(data), dataLen, nil)
	if err != nil {
		return fmt.Errorf("failed to upload: %w", err)
	}

	return nil
}

func (s *Storage) GetPrivateURL(key string, expires time.Duration) string {
	deadline := time.Now().Add(expires).Unix()
	privateURL := storage.MakePrivateURL(s.mac, s.domain, key, deadline)
	return privateURL
}

func (s *Storage) DownloadBytes(key string) ([]byte, error) {
	privateURL := s.GetPrivateURL(key, time.Hour)
	
	resp, err := http.Get(privateURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download: status code %d", resp.StatusCode)
	}
	
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	return data, nil
}
