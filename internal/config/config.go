package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	Database DatabaseConfig
}

type AppConfig struct {
	Name string
	Env  string
	URL  string
}

type HTTPConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func (c HTTPConfig) Addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

type DatabaseConfig struct {
	Driver   string
	DSN      string
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
}

func Load(paths ...string) (Config, error) {
	for _, path := range paths {
		if strings.TrimSpace(path) != "" {
			_ = godotenv.Load(path)
		}
	}

	cfg := Config{
		App: AppConfig{
			Name: env("APP_NAME", "Formora Taiwan"),
			Env:  env("APP_ENV", "development"),
			URL:  strings.TrimRight(env("APP_URL", "http://localhost:8080"), "/"),
		},
		HTTP: HTTPConfig{
			Host:            env("HTTP_HOST", "127.0.0.1"),
			Port:            envInt("HTTP_PORT", 8080),
			ReadTimeout:     envDuration("READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    envDuration("WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:     envDuration("IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: envDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		Database: DatabaseConfig{
			Driver:   strings.ToLower(env("DATABASE_DRIVER", "")),
			DSN:      env("DATABASE_DSN", ""),
			Host:     env("DATABASE_HOST", "127.0.0.1"),
			Port:     envInt("DATABASE_PORT", 5432),
			Name:     env("DATABASE_NAME", "university_agency"),
			User:     env("DATABASE_USER", "postgres"),
			Password: env("DATABASE_PASSWORD", "postgres"),
			SSLMode:  env("DATABASE_SSLMODE", "disable"),
		},
	}

	if cfg.HTTP.Port <= 0 || cfg.HTTP.Port > 65535 {
		return Config{}, fmt.Errorf("HTTP_PORT must be between 1 and 65535")
	}

	return cfg, nil
}

func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(value)
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value := env(key, "")
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := env(key, "")
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
