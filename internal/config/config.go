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

	dbURL := firstNonEmpty(
		os.Getenv("DATABASE_URL_POOLER"),
		os.Getenv("DB_URL_POOLER"),
		os.Getenv("DATABASE_URL_IPV4"),
		os.Getenv("DB_URL_IPV4"),
		os.Getenv("DATABASE_URL"),
		os.Getenv("DB_URL"),
	)

	cfg := &Config{
		DBURL:     dbURL,
		JWTSecret: os.Getenv("JWT_SECRET"),
		APIPort:   port,
	}

	return cfg, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}
