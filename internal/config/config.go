package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBURL                  string
	JWTSecret              string
	APIPort                string
	SupabaseURL            string
	SupabaseServiceRoleKey string
	SupabaseStorageBucket  string
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

	dbURL := os.Getenv("DATABASE_URL_POOLER")

	cfg := &Config{
		DBURL:                  dbURL,
		JWTSecret:              os.Getenv("JWT_SECRET"),
		APIPort:                port,
		SupabaseURL:            os.Getenv("SUPABASE_URL"),
		SupabaseServiceRoleKey: os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
		SupabaseStorageBucket:  getEnvOrDefault("SUPABASE_STORAGE_BUCKET", "time-records"),
	}

	return cfg, nil
}

func getEnvOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
