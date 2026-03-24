package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func Connect(dbURL string) (*sql.DB, error) {
	if strings.TrimSpace(dbURL) == "" {
		return nil, fmt.Errorf("database URL not provided (use DATABASE_URL_POOLER)")
	}

	dbURL, err := ensureSSLMode(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize database URL: %w", err)
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	for attempt := 1; attempt <= 10; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = db.PingContext(ctx)
		cancel()

		if err == nil {
			return db, nil
		}

		if attempt < 10 {
			time.Sleep(2 * time.Second)
		}
	}

	_ = db.Close()
	return nil, fmt.Errorf("failed to connect to database after multiple attempts: %w", err)
}

func ensureSSLMode(dbURL string) (string, error) {
	parsedURL, err := url.Parse(dbURL)
	if err != nil {
		return "", err
	}

	query := parsedURL.Query()
	if query.Get("sslmode") == "" {
		query.Set("sslmode", "require")
		parsedURL.RawQuery = query.Encode()
	}

	return parsedURL.String(), nil
}

func RunMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create database driver: %w", err)
	}

	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("could not create source driver: %w", err)
	}

	m, err := migrate.NewWithInstance(
		"iofs",
		source,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run up migrations: %w", err)
	}

	return nil
}
