package config

import "fmt"

// Postgres holds PostgreSQL connection settings.
type Postgres struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// LoadPostgresFromEnv reads standard DB_* environment variables.
func LoadPostgresFromEnv() Postgres {
	return Postgres{
		Host:     GetEnv("DB_HOST", "localhost"),
		Port:     GetEnv("DB_PORT", "5432"),
		User:     GetEnv("DB_USER", "postgres"),
		Password: GetEnv("DB_PASSWORD", ""),
		Name:     GetEnv("DB_NAME", "postgres"),
		SSLMode:  GetEnv("DB_SSLMODE", "disable"),
	}
}

// WithEnvDefaults fills SSLMode from the environment when unset.
func (p Postgres) WithEnvDefaults() Postgres {
	if p.SSLMode == "" {
		p.SSLMode = GetEnv("DB_SSLMODE", "disable")
	}
	return p
}

// DSN returns a libpq connection string for GORM/pgx.
func (p Postgres) DSN() string {
	p = p.WithEnvDefaults()
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.Name, p.SSLMode,
	)
}
