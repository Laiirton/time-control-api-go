package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBURL     string
	JWTSecret string
	APIPort   string
}

func Load() (*Config, error) {
	godotenv.Load()

	cfg := &Config{
		DBURL:     os.Getenv("DB_URL"),
		JWTSecret: os.Getenv("JWT_SECRET"),
		APIPort:   os.Getenv("API_PORT"),
	}

	if cfg.APIPort == "" {
		cfg.APIPort = "8080"
	}

	return cfg, nil
}
