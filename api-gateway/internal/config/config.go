package config

import (
	"os"
	"strconv"
	"strings"
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
	CORSAllowedOrigins    []string
	CookieSecure          bool
	CookieDomain          string
	AccessCookieName      string
	RefreshCookieName     string
	CSRFCookieName        string
	CSRFHeaderName        string
	RefreshCookieTTL      time.Duration
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
		CORSAllowedOrigins:    splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")),
		CookieSecure:          getBoolEnv("COOKIE_SECURE", false),
		CookieDomain:          strings.TrimSpace(getEnv("COOKIE_DOMAIN", "")),
		AccessCookieName:      getEnv("ACCESS_COOKIE_NAME", "berusmenu_access"),
		RefreshCookieName:     getEnv("REFRESH_COOKIE_NAME", "berusmenu_refresh"),
		CSRFCookieName:        getEnv("CSRF_COOKIE_NAME", "berusmenu_csrf"),
		CSRFHeaderName:        getEnv("CSRF_HEADER_NAME", "X-CSRF-Token"),
		RefreshCookieTTL:      getDurationEnv("REFRESH_COOKIE_TTL", 30*24*time.Hour),
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

func getBoolEnv(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	result, err := strconv.ParseBool(value)
	if err != nil {
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
