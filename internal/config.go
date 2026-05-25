package internal

import (
	"fmt"
	"os"
)

type Config struct {
	HTTPAddr       string
	DatabaseDSN    string
	LogLevel       string
	MigrationDir   string
}

func Load() (*Config, error) {
	cfg := &Config{
		HTTPAddr:     getEnv("HTTP_ADDR", ":8080"),
		DatabaseDSN:  getEnv("DATABASE_DSN", ""),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		MigrationDir: getEnv("MIGRATION_DIR", "migrations"),
	}
	if cfg.DatabaseDSN == "" {
		return nil, fmt.Errorf("DATABASE_DSN is required")
	}
	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
