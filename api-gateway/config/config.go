package config

import (
	"micro-warehouse/api-gateway/middleware"
	"os"
	"time"
)

func LoadJWTConfig() middleware.JWTConfig {
	secretKey := getEnv("JWT_SECRET_KEY", "your-secret-key-change-this-in-production")
	issuer := getEnv("JWT_ISSUER", "warehouse-api-gateway")
	durationStr := getEnv("JWT_DURATION", "1h")

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		// Default to 1 hour if parsing fails
		duration = 1 * time.Hour
	}

	return middleware.JWTConfig{
		SecretKey: secretKey,
		Issuer:    issuer,
		Duration:  duration,
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
