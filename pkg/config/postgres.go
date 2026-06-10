package config

import "fmt"

// Postgres holds PostgreSQL connection settings.
type Postgres struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

// DSN returns a libpq connection string for GORM/pgx.
func (p Postgres) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		p.Host, p.Port, p.User, p.Password, p.Name,
	)
}
