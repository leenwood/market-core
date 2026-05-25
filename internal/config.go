package internal

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTP     HTTPConfig
	Postgres PostgresConfig
	Log      LogConfig
	OTel     OTelConfig
	App      AppConfig
}

type HTTPConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	PprofEnabled bool
}

type PostgresConfig struct {
	DSN          string
	MigrationDir string
}

type LogConfig struct {
	Level  string
	Format string
}

type OTelConfig struct {
	Enabled     bool
	Exporter    string
	Endpoint    string
	ServiceName string
}

type AppConfig struct {
	Env string
}

func Load() (*Config, error) {
	dsn, err := requireEnv("DATABASE_DSN")
	if err != nil {
		return nil, err
	}
	return &Config{
		HTTP: HTTPConfig{
			Addr:         getEnv("HTTP_ADDR", ":8080"),
			ReadTimeout:  getEnvDuration("HTTP_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getEnvDuration("HTTP_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getEnvDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
			PprofEnabled: getEnvBool("HTTP_PPROF_ENABLED", false),
		},
		Postgres: PostgresConfig{
			DSN:          dsn,
			MigrationDir: getEnv("MIGRATION_DIR", "migrations"),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		OTel: OTelConfig{
			Enabled:     getEnvBool("OTEL_ENABLED", false),
			Exporter:    getEnv("OTEL_EXPORTER", "stdout"),
			Endpoint:    getEnv("OTEL_ENDPOINT", ""),
			ServiceName: getEnv("OTEL_SERVICE_NAME", "market-core"),
		},
		App: AppConfig{
			Env: getEnv("APP_ENV", "production"),
		},
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func getEnvStringSlice(key string, fallback []string) []string {
	if v := os.Getenv(key); v != "" {
		return strings.Split(v, ",")
	}
	return fallback
}

func requireEnv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("required environment variable %q is not set", key)
	}
	return v, nil
}
