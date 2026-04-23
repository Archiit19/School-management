package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	JWTSecret            string
	AuthServiceURL       string
	InternalServiceToken string
	Port                 string
}

func Load() *Config {
	return &Config{
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5437"),
		DBUser:               getEnv("DB_USER", "attendance_user"),
		DBPassword:           getEnv("DB_PASSWORD", "attendance_pass"),
		DBName:               getEnv("DB_NAME", "attendance_db"),
		JWTSecret:            getEnv("JWT_SECRET", "super-secret-jwt-key-change-in-production"),
		AuthServiceURL:       getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
		InternalServiceToken: getEnv("INTERNAL_SERVICE_TOKEN", ""),
		Port:                 getEnv("PORT", "8085"),
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
