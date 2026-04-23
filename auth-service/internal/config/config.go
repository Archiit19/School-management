package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBHost                string
	DBPort                string
	DBUser                string
	DBPassword            string
	DBName                string
	JWTSecret             string
	UserServiceURL        string
	InternalServiceToken  string
	Port                  string
}

func Load() *Config {
	return &Config{
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5433"),
		DBUser:               getEnv("DB_USER", "auth_user"),
		DBPassword:           getEnv("DB_PASSWORD", "auth_pass"),
		DBName:               getEnv("DB_NAME", "auth_db"),
		JWTSecret:            getEnv("JWT_SECRET", "super-secret-jwt-key-change-in-production"),
		UserServiceURL:       getEnv("USER_SERVICE_URL", "http://localhost:8082"),
		InternalServiceToken: getEnv("INTERNAL_SERVICE_TOKEN", ""),
		Port:                 getEnv("PORT", "8081"),
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
