package tracer

import (
	"strconv"
	"strings"

	pkgconfig "github.com/Archiit19/School-management/pkg/config"
)

// Config controls tracer provider construction.
type Config struct {
	Service     string
	Enabled     bool
	Endpoint    string
	Insecure    bool
	SampleRatio float64
}

// LoadConfigFromEnv reads tracer settings from the environment.
func LoadConfigFromEnv(service string) Config {
	ratio := 1.0
	if raw := pkgconfig.GetEnv("TRACE_SAMPLE_RATIO", ""); raw != "" {
		if parsed, err := strconv.ParseFloat(raw, 64); err == nil && parsed >= 0 && parsed <= 1 {
			ratio = parsed
		}
	}

	return Config{
		Service:     service,
		Enabled:     parseBool(pkgconfig.GetEnv("TRACE_ENABLED", "false")),
		Endpoint:    pkgconfig.GetEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4318"),
		Insecure:    parseBool(pkgconfig.GetEnv("OTEL_EXPORTER_OTLP_INSECURE", "true")),
		SampleRatio: ratio,
	}
}

func parseBool(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
