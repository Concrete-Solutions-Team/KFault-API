package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	JWTServer   string
	Port        string
}

func LoadConfig() *Config {
	godotenv.Load()

	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTServer:   os.Getenv("JWT_SECRET"),
		Port:        os.Getenv("PORT"),
	}

	return cfg
}
