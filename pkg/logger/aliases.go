package logger

import "github.com/Archiit19/School-management/pkg/logger/core"

// Public types and helpers — re-exported from core so services import pkg/logger only.

type (
	Logger = core.Logger
	Field  = core.Field
	Config = core.Config
	Backend = core.Backend
)

const (
	BackendSlog    = core.BackendSlog
	BackendZap     = core.BackendZap
	BackendZerolog = core.BackendZerolog
)

var (
	String   = core.String
	Int      = core.Int
	Int64    = core.Int64
	Bool     = core.Bool
	Any      = core.Any
	Err      = core.Err
	Duration = core.Duration
	Time     = core.Time
	KV       = core.KV
)

// New constructs a Logger for the configured backend.
func New(cfg Config) (Logger, error) {
	return core.New(cfg)
}

// LoadConfigFromEnv reads logger settings from the environment.
func LoadConfigFromEnv(service string) Config {
	return core.LoadConfigFromEnv(service)
}
