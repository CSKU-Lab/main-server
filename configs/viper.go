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
}

func NewConfig() *Config {
	var config Config

	viper.AutomaticEnv()
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")

	// Default values
	viper.SetDefault("PORT", "8080")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("❌ Error reading config file")
	}

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatal("❌ Unable to decode into struct")
	}

	if config.JWTSecret == "" {
		log.Fatal("❌ JWT_SECRET must be set")
	}

	return &config
}
