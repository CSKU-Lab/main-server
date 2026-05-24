package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ApiURL             string
	DatabaseURL        string
	Port               string
	GoogleClientID     string
	GoogleClientSecret string
	JWTSecret          string
	JWTRefreshSecret   string
	COOKIE_DOMAIN      string
	DevMode            bool
	FRONTEND_URL       string
	S3_AccessKeyID     string
	S3_SecretAccessKey string
	S3_UseSSL          bool
	S3_Endpoint        string
	S3_Frontend_URL    string
	S3_Bucket          string
	GRADER_SERVER_URL  string
	CONFIG_SERVER_URL  string
	TASK_SERVER_URL    string
	RBMQ_SERVER_URL    string
	REDIS_ADDR         string
	REDIS_PASSWORD     string
	REDIS_SERVER_URL   string
	InternalToken      string
}

func NewConfig() *Config {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Println("Error loading .env file:", err)
	}

	return &Config{
		ApiURL:             os.Getenv("API_URL"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		Port:               os.Getenv("PORT"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		JWTRefreshSecret:   os.Getenv("JWT_REFRESH_SECRET"),
		COOKIE_DOMAIN:      os.Getenv("COOKIE_DOMAIN"),
		DevMode:            os.Getenv("DEV_MODE") == "true",
		FRONTEND_URL:       os.Getenv("FRONTEND_URL"),
		S3_AccessKeyID:     os.Getenv("S3_ACCESS_KEY_ID"),
		S3_SecretAccessKey: os.Getenv("S3_SECRET_ACCESS_KEY"),
		S3_UseSSL:          os.Getenv("S3_USE_SSL") == "true",
		S3_Endpoint:        os.Getenv("S3_ENDPOINT"),
		S3_Frontend_URL:    os.Getenv("S3_FRONTEND_URL"),
		S3_Bucket:          os.Getenv("S3_BUCKET"),
		GRADER_SERVER_URL:  os.Getenv("GRADER_SERVER_URL"),
		CONFIG_SERVER_URL:  os.Getenv("CONFIG_SERVER_URL"),
		TASK_SERVER_URL:    os.Getenv("TASK_SERVER_URL"),
		RBMQ_SERVER_URL:    os.Getenv("RBMQ_SERVER_URL"),
		REDIS_ADDR:         os.Getenv("REDIS_ADDR"),
		REDIS_PASSWORD:     os.Getenv("REDIS_PASSWORD"),
		REDIS_SERVER_URL:   os.Getenv("REDIS_SERVER_URL"),
		InternalToken:      os.Getenv("INTERNAL_TOKEN"),
	}
}
