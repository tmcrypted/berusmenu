package config

import (
	"os"
	"strings"
)

type Config struct {
	AppEnv           string
	GRPCAddr         string
	PostgresDSN      string
	MatchingGRPCAddr string
	MigrationsDir    string
}

func Load() Config {
	return Config{
		AppEnv:           getEnv("APP_ENV", "local"),
		GRPCAddr:         getEnv("GRPC_ADDR", ":50052"),
		PostgresDSN:      getEnv("POSTGRES_DSN", "postgres://berusmenu:berusmenu@localhost:5432/berusmenu?sslmode=disable"),
		MatchingGRPCAddr: getEnv("MATCHING_GRPC_ADDR", "localhost:50053"),
		MigrationsDir:    getEnv("MIGRATIONS_DIR", "internal/migrations"),
	}
}

func getEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
