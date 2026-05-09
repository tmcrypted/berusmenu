package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv             string
	GRPCAddr           string
	PostgresDSN        string
	InventoryGRPCAddr  string
	AuthGRPCAddr       string
	KafkaBrokers       []string
	ResponseTopic      string
	MigrationsDir      string
	OutboxPollInterval time.Duration
	OutboxBatchSize    int
	OutboxRetryBackoff time.Duration
	OutboxClaimTimeout time.Duration
}

func Load() Config {
	return Config{
		AppEnv:             getEnv("APP_ENV", "local"),
		GRPCAddr:           getEnv("GRPC_ADDR", ":50053"),
		PostgresDSN:        getEnv("POSTGRES_DSN", "postgres://berusmenu:berusmenu@localhost:5432/berusmenu?sslmode=disable"),
		InventoryGRPCAddr:  getEnv("INVENTORY_GRPC_ADDR", "localhost:50052"),
		AuthGRPCAddr:       getEnv("AUTH_GRPC_ADDR", "localhost:50051"),
		KafkaBrokers:       splitCSV(getEnv("KAFKA_BROKERS", "localhost:9092")),
		ResponseTopic:      getEnv("MATCHING_RESPONSE_CREATED_TOPIC", "matching.response_created"),
		MigrationsDir:      getEnv("MIGRATIONS_DIR", "internal/migrations"),
		OutboxPollInterval: getDurationEnv("OUTBOX_POLL_INTERVAL", time.Second),
		OutboxBatchSize:    getIntEnv("OUTBOX_BATCH_SIZE", 50),
		OutboxRetryBackoff: getDurationEnv("OUTBOX_RETRY_BACKOFF", time.Second),
		OutboxClaimTimeout: getDurationEnv("OUTBOX_CLAIM_TIMEOUT", 5*time.Minute),
	}
}

func getEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func getIntEnv(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	result, err := strconv.Atoi(value)
	if err != nil || result < 1 {
		return fallback
	}
	return result
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	result, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return result
}
