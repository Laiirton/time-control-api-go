package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"net"
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
		return nil, fmt.Errorf("URL do banco não informada (use DATABASE_URL_POOLER/DB_URL_POOLER, DATABASE_URL_IPV4/DB_URL_IPV4, DATABASE_URL ou DB_URL)")
	}

	dbURL, err := ensureSSLMode(dbURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao normalizar URL do banco: %w", err)
	}

	dbURL, err = ensureIPv4HostAddr(dbURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao preparar conexão do banco: %w", err)
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir conexão: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	// Retry curto para reduzir falhas transitórias no boot do Render.
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
	if isSupabaseDirectHost(dbURL) && strings.Contains(strings.ToLower(err.Error()), "network is unreachable") {
		return nil, fmt.Errorf("erro ao conectar ao banco após múltiplas tentativas: %w. O host direto do Supabase pode exigir IPv6; no Render use a URL de pooler IPv4 em DATABASE_URL_POOLER", err)
	}

	return nil, fmt.Errorf("erro ao conectar ao banco após múltiplas tentativas: %w", err)
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

func ensureIPv4HostAddr(dbURL string) (string, error) {
	parsedURL, err := url.Parse(dbURL)
	if err != nil {
		return "", err
	}

	query := parsedURL.Query()
	if query.Get("hostaddr") != "" {
		return parsedURL.String(), nil
	}

	hostname := parsedURL.Hostname()
	if strings.TrimSpace(hostname) == "" {
		return parsedURL.String(), nil
	}

	if parsedIP := net.ParseIP(hostname); parsedIP != nil {
		if parsedIP.To4() == nil {
			return parsedURL.String(), nil
		}
		return parsedURL.String(), nil
	}

	ips, err := net.LookupIP(hostname)
	if err != nil {
		return parsedURL.String(), nil
	}

	for _, ip := range ips {
		if ip4 := ip.To4(); ip4 != nil {
			query.Set("hostaddr", ip4.String())
			parsedURL.RawQuery = query.Encode()
			return parsedURL.String(), nil
		}
	}

	return parsedURL.String(), nil
}

func isSupabaseDirectHost(dbURL string) bool {
	parsedURL, err := url.Parse(dbURL)
	if err != nil {
		return false
	}

	hostname := strings.ToLower(parsedURL.Hostname())
	return strings.HasPrefix(hostname, "db.") && strings.HasSuffix(hostname, ".supabase.co")
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
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run up migrations: %w", err)
	}

	return nil
}
