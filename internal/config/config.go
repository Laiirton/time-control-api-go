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

	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("API_PORT")
	}
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DB_URL")
	}

	cfg := &Config{
		DBURL:     dbURL,
		JWTSecret: os.Getenv("JWT_SECRET"),
		APIPort:   port,
	}

	return cfg, nil
}
