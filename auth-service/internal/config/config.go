package config

import pkgconfig "github.com/Archiit19/School-management/pkg/config"

type Config struct {
	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	JWTSecret            string
	UserServiceURL       string
	StudentServiceURL    string
	SchoolServiceURL     string
	InternalServiceToken string
	Port                 string
}

func Load() *Config {
	return &Config{
		DBHost:               pkgconfig.GetEnv("DB_HOST", "localhost"),
		DBPort:               pkgconfig.GetEnv("DB_PORT", "15433"),
		DBUser:               pkgconfig.GetEnv("DB_USER", "auth_user"),
		DBPassword:           pkgconfig.GetEnv("DB_PASSWORD", "auth_pass"),
		DBName:               pkgconfig.GetEnv("DB_NAME", "auth_db"),
		JWTSecret:            pkgconfig.GetEnv("JWT_SECRET", "super-secret-jwt-key-change-in-production"),
		UserServiceURL:       pkgconfig.GetEnv("USER_SERVICE_URL", "http://localhost:8082"),
		StudentServiceURL:    pkgconfig.GetEnv("STUDENT_SERVICE_URL", "http://localhost:8084"),
		SchoolServiceURL:     pkgconfig.GetEnv("SCHOOL_SERVICE_URL", "http://localhost:8088"),
		InternalServiceToken: pkgconfig.GetEnv("INTERNAL_SERVICE_TOKEN", ""),
		Port:                 pkgconfig.GetEnv("PORT", "8081"),
	}
}

func (c *Config) DSN() string {
	return pkgconfig.Postgres{
		Host:     c.DBHost,
		Port:     c.DBPort,
		User:     c.DBUser,
		Password: c.DBPassword,
		Name:     c.DBName,
	}.DSN()
}
