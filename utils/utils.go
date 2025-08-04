package utils

import (
	"log"
	"os"
	"strconv"
	"time"
)

func getEnvDefault(key string, a []string) (string, bool) {
	v := os.Getenv(key)
	if v != "" {
		return v, true
	}
	if len(a) > 0 {
		return a[0], true
	}
	return "", false
}

func GetEnv(key string, a ...string) string {
	if v, ok := getEnvDefault(key, a); ok {
		return v
	}
	log.Fatalf("Missing required string environment variable: %s", key)
	return ""
}

func GetEnvBool(key string, a ...string) bool {
	if v, ok := getEnvDefault(key, a); ok {
		if val, err := strconv.ParseBool(v); err == nil {
			return val
		}
	}
	log.Fatalf("Missing required bool environment variable: %s", key)
	return false
}

func GetEnvInt(key string, a ...string) int {
	if v, ok := getEnvDefault(key, a); ok {
		if intVal, err := strconv.Atoi(v); err == nil {
			return intVal
		}
	}
	log.Fatalf("Missing required int environment variable: %s", key)
	return 0
}

func GetEnvFloat(key string, a ...string) float64 {
	if v, ok := getEnvDefault(key, a); ok {
		if val, err := strconv.ParseFloat(v, 64); err == nil {
			return val
		}
	}
	log.Fatalf("Missing required float environment variable: %s", key)
	return 0
}

func GetEnvDuration(key string, a ...string) time.Duration {
	if v, ok := getEnvDefault(key, a); ok {
		if duration, err := time.ParseDuration(v); err == nil {
			return duration
		}
	}
	log.Fatalf("Missing required duration environment variable: %s", key)
	return 0
}
