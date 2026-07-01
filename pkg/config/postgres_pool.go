package config

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

// PostgresPool holds sql.DB pool tuning for production workloads.
type PostgresPool struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultPostgresPool returns conservative pool defaults.
func DefaultPostgresPool() PostgresPool {
	return PostgresPool{
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 15 * time.Minute,
	}
}

// LoadPostgresPoolFromEnv reads DB pool settings from the environment.
func LoadPostgresPoolFromEnv() PostgresPool {
	def := DefaultPostgresPool()
	return PostgresPool{
		MaxOpenConns:    GetEnvInt("DB_MAX_OPEN_CONNS", def.MaxOpenConns),
		MaxIdleConns:    GetEnvInt("DB_MAX_IDLE_CONNS", def.MaxIdleConns),
		ConnMaxLifetime: GetEnvDuration("DB_CONN_MAX_LIFETIME", def.ConnMaxLifetime),
		ConnMaxIdleTime: GetEnvDuration("DB_CONN_MAX_IDLE_TIME", def.ConnMaxIdleTime),
	}
}

// ConfigurePool applies pool settings to a GORM database handle.
func ConfigurePool(db *gorm.DB, pool PostgresPool) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	applyPool(sqlDB, pool)
	return nil
}

func applyPool(sqlDB *sql.DB, pool PostgresPool) {
	if pool.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(pool.MaxOpenConns)
	}
	if pool.MaxIdleConns >= 0 {
		sqlDB.SetMaxIdleConns(pool.MaxIdleConns)
	}
	if pool.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(pool.ConnMaxLifetime)
	}
	if pool.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(pool.ConnMaxIdleTime)
	}
}
