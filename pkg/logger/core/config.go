package core

import (
	"io"
	"os"
	"strings"

	pkgconfig "github.com/Archiit19/School-management/pkg/config"
)

// Config controls logger construction.
type Config struct {
	Service string
	Level   string
	Format  string
	Backend Backend
	Output  io.Writer
}

// LoadConfigFromEnv reads logger settings from the environment.
func LoadConfigFromEnv(service string) Config {
	return Config{
		Service: service,
		Level:   pkgconfig.GetEnv("LOG_LEVEL", "info"),
		Format:  pkgconfig.GetEnv("LOG_FORMAT", "text"),
		Backend: ParseBackend(pkgconfig.GetEnv("LOG_BACKEND", string(BackendSlog))),
		Output:  os.Stdout,
	}
}

// ParseBackend normalizes a backend name.
func ParseBackend(raw string) Backend {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(BackendZap):
		return BackendZap
	case string(BackendZerolog):
		return BackendZerolog
	default:
		return BackendSlog
	}
}

// ParseLevel normalizes a level name.
func ParseLevel(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return "debug"
	case "warn", "warning":
		return "warn"
	case "error":
		return "error"
	default:
		return "info"
	}
}
