package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"university_agency/internal/config"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(cfg config.DatabaseConfig) (*sql.DB, error) {
	driver := strings.ToLower(strings.TrimSpace(cfg.Driver))
	if driver == "" {
		return nil, fmt.Errorf("DATABASE_DRIVER is required")
	}

	sqlDriver, dsn, err := normalize(driver, cfg)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(sqlDriver, dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(12)
	db.SetMaxIdleConns(6)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}

func normalize(driver string, cfg config.DatabaseConfig) (string, string, error) {
	if cfg.DSN != "" {
		switch driver {
		case "postgres", "postgresql", "pgx":
			return "pgx", cfg.DSN, nil
		case "mysql", "mariadb":
			return "mysql", cfg.DSN, nil
		default:
			return "", "", fmt.Errorf("unsupported database driver %q", driver)
		}
	}

	switch driver {
	case "postgres", "postgresql", "pgx":
		query := url.Values{}
		query.Set("sslmode", cfg.SSLMode)
		dsn := url.URL{
			Scheme:   "postgres",
			User:     url.UserPassword(cfg.User, cfg.Password),
			Host:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Path:     cfg.Name,
			RawQuery: query.Encode(),
		}
		return "pgx", dsn.String(), nil
	case "mysql", "mariadb":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4,utf8",
			cfg.User,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Name,
		)
		return "mysql", dsn, nil
	default:
		return "", "", fmt.Errorf("unsupported database driver %q", driver)
	}
}
