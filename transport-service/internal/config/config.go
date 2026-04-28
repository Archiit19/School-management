package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
	DBName    string
	JWTSecret string
	Port      string
}

func Load() *Config {
	return &Config{
		DBHost:    getEnv("DB_HOST", "localhost"),
		DBPort:    getEnv("DB_PORT", "5432"),
		DBUser:    getEnv("DB_USER", "transport_user"),
		DBPass:    getEnv("DB_PASSWORD", "transport_pass"),
		DBName:    getEnv("DB_NAME", "transport_db"),
		JWTSecret: getEnv("JWT_SECRET", "super-secret-jwt-key-change-in-production"),
		Port:      getEnv("PORT", "8088"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPass, c.DBName,
	)
}
