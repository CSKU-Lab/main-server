package configs

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	ApiURL             string `mapstructure:"API_URL"`
	DatabaseURL        string `mapstructure:"DATABASE_URL"`
	Port               string `mapstructure:"PORT"`
	GoogleClientID     string `mapstructure:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `mapstructure:"GOOGLE_CLIENT_SECRET"`
	JWTSecret          string `mapstructure:"JWT_SECRET"`
	JWTRefreshSecret   string `mapstructure:"JWT_REFRESH_SECRET"`
	COOKIE_DOMAIN      string `mapstructure:"COOKIE_DOMAIN"`
	DEV_MODE           bool   `mapstructure:"DEV_MODE"`
	FRONTEND_URL       string `mapstructure:"FRONTEND_URL"`
	MinIO              MinIO
}

type MinIO struct {
	Endpoint        string `mapstructure:"MINIO_ENDPOINT"`
	AccessKeyID     string `mapstructure:"MINIO_ACCESS_KEY_ID"`
	SecretAccessKey string `mapstructure:"MINIO_SECRET_ACCESS_KEY"`
	UseSSL          bool   `mapstructure:"MINIO_USE_SSL"`
	Bucket          string `mapstructure:"MINIO_BUCKET"`
}

func NewConfig() *Config {
	var config Config

	viper.AutomaticEnv()
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")

	// Default values
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("DEV_MODE", false)

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("❌ Error reading config file")
	}

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatal("❌ Unable to decode into struct")
	}

	if config.JWTSecret == "" {
		log.Fatal("❌ JWT_SECRET must be set")
	}

	if config.GoogleClientID == "" {
		log.Fatal("❌ GOOGLE_CLIENT_ID must be set")
	}

	if config.GoogleClientSecret == "" {
		log.Fatal("❌ GOOGLE_CLIENT_SECRET must be set")
	}

	return &config
}
