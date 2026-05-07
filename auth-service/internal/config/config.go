package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv            string
	GRPCAddr          string
	PostgresDSN       string
	RedisAddr         string
	KafkaBrokers      []string
	VerificationTopic string
	JWTSecret         string
	AccessTTL         time.Duration
	RefreshTTL        time.Duration
	CodeTTL           time.Duration
	ResendCooldown    time.Duration
	LockoutTTL        time.Duration
	MaxVerifyAttempts int
	MigrationsDir     string
}

func Load() Config {
	return Config{
		AppEnv:            getEnv("APP_ENV", "local"),
		GRPCAddr:          getEnv("GRPC_ADDR", ":50051"),
		PostgresDSN:       getEnv("POSTGRES_DSN", "postgres://berusmenu:berusmenu@localhost:5432/berusmenu?sslmode=disable"),
		RedisAddr:         getEnv("REDIS_ADDR", "localhost:6379"),
		KafkaBrokers:      splitCSV(getEnv("KAFKA_BROKERS", "localhost:9092")),
		VerificationTopic: getEnv("AUTH_EMAIL_VERIFICATION_TOPIC", "auth.email_verification_requested"),
		JWTSecret:         getEnv("JWT_SECRET", "local-dev-secret-change-me"),
		AccessTTL:         getDurationEnv("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTTL:        getDurationEnv("REFRESH_TOKEN_TTL", 30*24*time.Hour),
		CodeTTL:           getDurationEnv("EMAIL_VERIFY_CODE_TTL", 10*time.Minute),
		ResendCooldown:    getDurationEnv("EMAIL_VERIFY_RESEND_COOLDOWN", time.Minute),
		LockoutTTL:        getDurationEnv("EMAIL_VERIFY_LOCKOUT_TTL", time.Hour),
		MaxVerifyAttempts: getIntEnv("EMAIL_VERIFY_MAX_ATTEMPTS", 3),
		MigrationsDir:     getEnv("MIGRATIONS_DIR", "internal/migrations"),
	}
}

func getEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return duration
}

func getIntEnv(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	result, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	if result <= 0 {
		return fallback
	}
	return result
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
