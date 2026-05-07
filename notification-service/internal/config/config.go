package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	KafkaBrokers      []string
	VerificationTopic string
	ResponseTopic     string
	ConsumerGroupID   string
	RetryAttempts     int
	RetryBackoff      time.Duration
}

func Load() Config {
	return Config{
		KafkaBrokers:      splitCSV(getEnv("KAFKA_BROKERS", "localhost:9092")),
		VerificationTopic: getEnv("AUTH_EMAIL_VERIFICATION_TOPIC", "auth.email_verification_requested"),
		ResponseTopic:     getEnv("MATCHING_RESPONSE_CREATED_TOPIC", "matching.response_created"),
		ConsumerGroupID:   getEnv("KAFKA_CONSUMER_GROUP_ID", "notification-service-group"),
		RetryAttempts:     getIntEnv("NOTIFICATION_RETRY_ATTEMPTS", 3),
		RetryBackoff:      getDurationEnv("NOTIFICATION_RETRY_BACKOFF", time.Second),
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
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
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
