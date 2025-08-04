package utils

import (
	"log"
	"os"
	"strconv"
	"time"
)

func GetEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func GetEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if val, err := strconv.ParseBool(v); err == nil {
			return val
		}
	}
	return fallback
}

func GetEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if intVal, err := strconv.Atoi(v); err == nil {
			return intVal
		}
	}
	return fallback
}

func GetEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if val, err := strconv.ParseFloat(v, 64); err == nil {
			return val
		}
	}
	return fallback
}

func GetEnvDuration(key string, fallback string) time.Duration {
	duration, err := time.ParseDuration(GetEnv(key, fallback))
	if err != nil {
		log.Fatalf("Invalid %s: %v", key, err)
	}
	return duration
}
