package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	JWTSecret         string
	Port              string
	AuthServiceURL    string
	StudentServiceURL string
}

func Load() *Config {
	return &Config{
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnv("DB_PORT", "5435"),
		DBUser:            getEnv("DB_USER", "academic_user"),
		DBPassword:        getEnv("DB_PASSWORD", "academic_pass"),
		DBName:            getEnv("DB_NAME", "academic_db"),
		JWTSecret:         getEnv("JWT_SECRET", "super-secret-jwt-key-change-in-production"),
		Port:              getEnv("PORT", "8083"),
		AuthServiceURL:    getEnv("AUTH_SERVICE_URL", "http://auth-service:8081"),
		StudentServiceURL: getEnv("STUDENT_SERVICE_URL", "http://student-service:8084"),
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
