package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL        string
	JWTServer          string
	Port               string
	FrontendURL        string
	CloudflareAPIToken string
	S3AccountID        string
	S3AccessKeyID      string
	S3AccessKeySecret  string
	S3BucketName       string
}

func LoadConfig() *Config {
	godotenv.Load()

	cfg := &Config{
		DatabaseURL:        getDatabaseURL(),
		JWTServer:          os.Getenv("JWT_SECRET"),
		Port:               os.Getenv("PORT"),
		FrontendURL:        os.Getenv("FRONTEND_URL"),
		CloudflareAPIToken: os.Getenv("CLOUDFLARE_API_TOKEN"),
		S3AccountID:        os.Getenv("S3_ACCOUNT_ID"),
		S3AccessKeyID:      os.Getenv("S3_ACCESS_KEY_ID"),
		S3AccessKeySecret:  os.Getenv("S3_ACCESS_KEY_SECRET"),
		S3BucketName:       os.Getenv("S3_BUCKET_NAME"),
	}

	return cfg
}

func getDatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s&channel_binding=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
		os.Getenv("DB_CHANNEL_BINDING"),
	)
}
