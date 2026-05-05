package config

import (
	"os"
	"time"
)

type Config struct {
	HTTPAddr              string
	AuthGRPCAddr          string
	InventoryGRPCAddr     string
	MatchingGRPCAddr      string
	CommunicationGRPCAddr string
	RequestTimeout        time.Duration
	AppEnv                string
}

func Load() Config {
	return Config{
		HTTPAddr:              getEnv("HTTP_ADDR", ":8080"),
		AuthGRPCAddr:          getEnv("AUTH_GRPC_ADDR", "localhost:50051"),
		InventoryGRPCAddr:     getEnv("INVENTORY_GRPC_ADDR", "localhost:50052"),
		MatchingGRPCAddr:      getEnv("MATCHING_GRPC_ADDR", "localhost:50053"),
		CommunicationGRPCAddr: getEnv("COMMUNICATION_GRPC_ADDR", "localhost:50054"),
		RequestTimeout:        getDurationEnv("REQUEST_TIMEOUT", 5*time.Second),
		AppEnv:                getEnv("APP_ENV", "local"),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return duration
}
