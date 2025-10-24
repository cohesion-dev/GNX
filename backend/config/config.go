package config

import (
	"os"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	OpenAI   OpenAIConfig
	Qiniu    QiniuConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type ServerConfig struct {
	Port string
}

type OpenAIConfig struct {
	APIKey string
}

type QiniuConfig struct {
	AccessKey string
	SecretKey string
	Bucket    string
	Domain    string
}

func LoadConfig() (*Config, error) {
	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "gnx"),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		OpenAI: OpenAIConfig{
			APIKey: getEnv("OPENAI_API_KEY", ""),
		},
		Qiniu: QiniuConfig{
			AccessKey: getEnv("QINIU_ACCESS_KEY", ""),
			SecretKey: getEnv("QINIU_SECRET_KEY", ""),
			Bucket:    getEnv("QINIU_BUCKET", ""),
			Domain:    getEnv("QINIU_DOMAIN", ""),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
