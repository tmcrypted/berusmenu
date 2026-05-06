package config

import (
	"os"
	"strings"
)

type Config struct {
	KafkaBrokers      []string
	VerificationTopic string
	ConsumerGroupID   string
}

func Load() Config {
	return Config{
		KafkaBrokers:      splitCSV(getEnv("KAFKA_BROKERS", "localhost:9092")),
		VerificationTopic: getEnv("AUTH_EMAIL_VERIFICATION_TOPIC", "auth.email_verification_requested"),
		ConsumerGroupID:   getEnv("KAFKA_CONSUMER_GROUP_ID", "notification-service-group"),
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
