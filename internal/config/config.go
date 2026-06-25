package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL        string
	JWTServer          string
	Port               string
	CloudflareAPIToken string
	S3AccountID        string
	S3AccessKeyID      string
	S3AccessKeySecret  string
	S3BucketName       string
}

func LoadConfig() *Config {
	godotenv.Load()

	cfg := &Config{
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		JWTServer:          os.Getenv("JWT_SECRET"),
		Port:               os.Getenv("PORT"),
		CloudflareAPIToken: os.Getenv("CLOUDFLARE_API_TOKEN"),
		S3AccountID:        os.Getenv("S3_ACCOUNT_ID"),
		S3AccessKeyID:      os.Getenv("S3_ACCESS_KEY_ID"),
		S3AccessKeySecret:  os.Getenv("S3_ACCESS_KEY_SECRET"),
		S3BucketName:       os.Getenv("S3_BUCKET_NAME"),
	}

	return cfg
}
