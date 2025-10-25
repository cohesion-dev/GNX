package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Server   ServerConfig   `mapstructure:"server"`
	OpenAI   OpenAIConfig   `mapstructure:"openai"`
	Qiniu    QiniuConfig    `mapstructure:"qiniu"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type OpenAIConfig struct {
	APIKey string `mapstructure:"api_key"`
}

type QiniuConfig struct {
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	Domain    string `mapstructure:"domain"`
}

func LoadConfig() (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", "5432")
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "")
	v.SetDefault("database.dbname", "gnx")
	v.SetDefault("server.port", "8080")

	v.AutomaticEnv()
	v.SetEnvPrefix("")

	v.BindEnv("database.host", "DB_HOST")
	v.BindEnv("database.port", "DB_PORT")
	v.BindEnv("database.user", "DB_USER")
	v.BindEnv("database.password", "DB_PASSWORD")
	v.BindEnv("database.dbname", "DB_NAME")
	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("openai.api_key", "OPENAI_API_KEY")
	v.BindEnv("qiniu.access_key", "QINIU_ACCESS_KEY")
	v.BindEnv("qiniu.secret_key", "QINIU_SECRET_KEY")
	v.BindEnv("qiniu.bucket", "QINIU_BUCKET")
	v.BindEnv("qiniu.domain", "QINIU_DOMAIN")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
