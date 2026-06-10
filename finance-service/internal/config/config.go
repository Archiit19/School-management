package config

import pkgconfig "github.com/Archiit19/School-management/pkg/config"

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret            string
	Port                 string
	UserServiceURL       string
	InternalServiceToken string
}

func Load() *Config {
	return &Config{
		DBHost:     pkgconfig.GetEnv("DB_HOST", "localhost"),
		DBPort:     pkgconfig.GetEnv("DB_PORT", "5439"),
		DBUser:     pkgconfig.GetEnv("DB_USER", "finance_user"),
		DBPassword: pkgconfig.GetEnv("DB_PASSWORD", "finance_pass"),
		DBName:     pkgconfig.GetEnv("DB_NAME", "finance_db"),
		JWTSecret:            pkgconfig.GetEnv("JWT_SECRET", "super-secret-jwt-key-change-in-production"),
		Port:                 pkgconfig.GetEnv("PORT", "8087"),
		UserServiceURL:       pkgconfig.GetEnv("USER_SERVICE_URL", "http://user-service:8082"),
		InternalServiceToken: pkgconfig.GetEnv("INTERNAL_SERVICE_TOKEN", "dev-internal-token-change-in-production"),
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
