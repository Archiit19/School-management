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
	AcademicServiceURL   string
	InternalServiceToken string
	Port                 string
}

func Load() *Config {
	return &Config{
		DBHost:               pkgconfig.GetEnv("DB_HOST", "localhost"),
		DBPort:               pkgconfig.GetEnv("DB_PORT", "5437"),
		DBUser:               pkgconfig.GetEnv("DB_USER", "attendance_user"),
		DBPassword:           pkgconfig.GetEnv("DB_PASSWORD", "attendance_pass"),
		DBName:               pkgconfig.GetEnv("DB_NAME", "attendance_db"),
		JWTSecret:            pkgconfig.GetEnv("JWT_SECRET", "super-secret-jwt-key-change-in-production"),
		UserServiceURL:       pkgconfig.GetEnv("USER_SERVICE_URL", "http://user-service:8082"),
		AcademicServiceURL:   pkgconfig.GetEnv("ACADEMIC_SERVICE_URL", "http://academic-service:8083"),
		InternalServiceToken: pkgconfig.GetEnv("INTERNAL_SERVICE_TOKEN", ""),
		Port:                 pkgconfig.GetEnv("PORT", "8085"),
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
