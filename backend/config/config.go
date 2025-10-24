package config

import (
	"os"
)

type Config struct {
	ServerPort   string
	DatabaseURL  string
	OpenAIAPIKey string
	QiniuConfig  QiniuConfig
}

type QiniuConfig struct {
	AccessKey string
	SecretKey string
	Bucket    string
	Domain    string
}

func LoadConfig() (*Config, error) {
	return &Config{
		ServerPort:   getEnv("SERVER_PORT", "8080"),
		DatabaseURL:  getEnv("DATABASE_URL", ""),
		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),
		QiniuConfig: QiniuConfig{
			AccessKey: getEnv("QINIU_ACCESS_KEY", ""),
			SecretKey: getEnv("QINIU_SECRET_KEY", ""),
			Bucket:    getEnv("QINIU_BUCKET", ""),
			Domain:    getEnv("QINIU_DOMAIN", ""),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
