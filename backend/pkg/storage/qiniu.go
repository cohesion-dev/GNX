package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type QiniuClient struct {
	accessKey string
	secretKey string
	bucket    string
	domain    string
}

func NewQiniuClient(accessKey, secretKey, bucket, domain string) *QiniuClient {
	return &QiniuClient{
		accessKey: accessKey,
		secretKey: secretKey,
		bucket:    bucket,
		domain:    domain,
	}
}

func (c *QiniuClient) UploadImage(key string, data []byte) (string, error) {
	return c.upload(key, data, "image/png")
}

func (c *QiniuClient) UploadAudio(key string, data []byte) (string, error) {
	return c.upload(key, data, "audio/mpeg")
}

func (c *QiniuClient) UploadFile(key string, data []byte, contentType string) (string, error) {
	return c.upload(key, data, contentType)
}

func (c *QiniuClient) upload(key string, data []byte, contentType string) (string, error) {
	uploadToken := c.generateUploadToken()

	url := fmt.Sprintf("https://upload.qiniup.com/putb64/-1/key/%s", key)

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", fmt.Sprintf("UpToken %s", uploadToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed: %s", string(body))
	}

	return fmt.Sprintf("https://%s/%s", c.domain, key), nil
}

func (c *QiniuClient) generateUploadToken() string {
	return fmt.Sprintf("%s:%s:%s", c.accessKey, c.secretKey, c.bucket)
}

func (c *QiniuClient) GenerateTTS(text, voice string) ([]byte, error) {
	requestBody, _ := json.Marshal(map[string]string{
		"text":  text,
		"voice": voice,
	})

	req, err := http.NewRequest("POST", "https://tts-api.qiniu.com/v1/tts", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Qiniu %s:%s", c.accessKey, c.secretKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TTS generation failed: %s", string(body))
	}

	return io.ReadAll(resp.Body)
}
