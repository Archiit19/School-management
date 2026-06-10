package config

import pkgconfig "github.com/Archiit19/School-management/pkg/config"

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
		DBHost:               pkgconfig.GetEnv("DB_HOST", "localhost"),
		DBPort:               pkgconfig.GetEnv("DB_PORT", "5440"),
		DBUser:               pkgconfig.GetEnv("DB_USER", "school_user"),
		DBPassword:           pkgconfig.GetEnv("DB_PASSWORD", "school_pass"),
		DBName:               pkgconfig.GetEnv("DB_NAME", "school_db"),
		JWTSecret:            pkgconfig.GetEnv("JWT_SECRET", "super-secret-jwt-key-change-in-production"),
		AuthServiceURL:       pkgconfig.GetEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
		InternalServiceToken: pkgconfig.GetEnv("INTERNAL_SERVICE_TOKEN", ""),
		Port:                 pkgconfig.GetEnv("PORT", "8088"),
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
