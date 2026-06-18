package config

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// GORMOptions optional settings when opening PostgreSQL via GORM.
type GORMOptions struct {
	Config *gorm.Config
	Pool   *PostgresPool
}

// OpenGORM opens PostgreSQL with GORM, configures the connection pool, and pings the database.
func OpenGORM(dsn string, opts *GORMOptions) (*gorm.DB, error) {
	gormCfg := &gorm.Config{}
	if opts != nil && opts.Config != nil {
		gormCfg = opts.Config
	}

	db, err := gorm.Open(postgres.Open(dsn), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("postgres: open: %w", err)
	}

	pool := LoadPostgresPoolFromEnv()
	if opts != nil && opts.Pool != nil {
		pool = *opts.Pool
	}
	if err := ConfigurePool(db, pool); err != nil {
		return nil, fmt.Errorf("postgres: configure pool: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("postgres: sql handle: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	return db, nil
}
