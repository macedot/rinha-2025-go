package utils

import (
	"log"
	"os"
	"time"
)

func GetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s environment variable not set", key)
	}
	return value
}

func GetEnvOr(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func GetEnvDurationOr(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(valueStr)
	if err != nil {
		log.Printf("warning: invalid duration for %s: %s, using default", key, valueStr)
		return defaultValue
	}
	return duration
}
