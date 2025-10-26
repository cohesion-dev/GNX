package config

import (
	"os"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Storage  StorageConfig
	AI       AIConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type StorageConfig struct {
	AccessKey string
	SecretKey string
	Bucket    string
	Domain    string
}

type AIConfig struct {
	APIKey        string
	BaseURL       string
	ImageModel    string
	LanguageModel string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "gnx"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Storage: StorageConfig{
			AccessKey: getEnv("QINIU_ACCESS_KEY", ""),
			SecretKey: getEnv("QINIU_SECRET_KEY", ""),
			Bucket:    getEnv("QINIU_BUCKET", ""),
			Domain:    getEnv("QINIU_DOMAIN", ""),
		},
		AI: AIConfig{
			APIKey:        getEnv("OPENAI_API_KEY", ""),
			BaseURL:       getEnv("OPENAI_BASE_URL", "https://openai.qiniu.com/v1"),
			ImageModel:    getEnv("OPENAI_IMAGE_MODEL", "gemini-2.5-flash-image"),
			LanguageModel: getEnv("OPENAI_LANGUAGE_MODEL", "deepseek/deepseek-v3.1-terminus"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
