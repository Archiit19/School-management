package config

import "os"

// GetEnv returns the environment variable value or fallback when unset/empty.
func GetEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
