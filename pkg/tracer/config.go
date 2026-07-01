package tracer

import (
	"os"
	"strconv"
	"strings"
	"time"

	pkgconfig "github.com/Archiit19/School-management/pkg/config"
)

const defaultSampleRatio = 0.1

// Config controls tracer provider construction.
type Config struct {
	Service     string
	Version     string
	Environment string
	Enabled     bool
	Endpoint    string
	Insecure    bool
	SampleRatio float64

	BatchTimeout       time.Duration
	MaxExportBatchSize int
	MaxQueueSize       int
}

// LoadConfigFromEnv reads tracer settings from the environment.
func LoadConfigFromEnv(service string) Config {
	endpoint, insecure := loadOTLPEndpoint()
	return Config{
		Service:            service,
		Version:            pkgconfig.GetEnv("SERVICE_VERSION", ""),
		Environment:        loadEnvironment(),
		Enabled:            loadEnabled(),
		Endpoint:           endpoint,
		Insecure:           insecure,
		SampleRatio:        loadSampleRatio(),
		BatchTimeout:       pkgconfig.GetEnvDuration("OTEL_BSP_SCHEDULE_DELAY", 5*time.Second),
		MaxExportBatchSize: pkgconfig.GetEnvInt("OTEL_BSP_MAX_EXPORT_BATCH_SIZE", 512),
		MaxQueueSize:       pkgconfig.GetEnvInt("OTEL_BSP_MAX_QUEUE_SIZE", 2048),
	}
}

func loadEnabled() bool {
	if raw := os.Getenv("TRACE_ENABLED"); raw != "" {
		return parseBool(raw)
	}
	if os.Getenv("OTEL_SDK_DISABLED") != "" {
		return !pkgconfig.GetEnvBool("OTEL_SDK_DISABLED", false)
	}
	switch strings.ToLower(pkgconfig.GetEnv("OTEL_TRACES_EXPORTER", "")) {
	case "otlp", "otlp/http", "otlp_http":
		return true
	case "none":
		return false
	}
	return false
}

func loadSampleRatio() float64 {
	if raw := pkgconfig.GetEnv("TRACE_SAMPLE_RATIO", ""); raw != "" {
		if parsed, err := strconv.ParseFloat(raw, 64); err == nil && parsed >= 0 && parsed <= 1 {
			return parsed
		}
	}
	if raw := pkgconfig.GetEnv("OTEL_TRACES_SAMPLER_ARG", ""); raw != "" {
		if parsed, err := strconv.ParseFloat(raw, 64); err == nil && parsed >= 0 && parsed <= 1 {
			return parsed
		}
	}
	return defaultSampleRatio
}

func loadEnvironment() string {
	if env := pkgconfig.GetEnv("DEPLOYMENT_ENVIRONMENT", ""); env != "" {
		return env
	}
	return pkgconfig.GetEnv("ENV", "")
}

func loadOTLPEndpoint() (host string, insecure bool) {
	raw := pkgconfig.GetEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4318")
	insecure = pkgconfig.GetEnvBool("OTEL_EXPORTER_OTLP_INSECURE", false)

	switch {
	case strings.HasPrefix(raw, "http://"):
		raw = strings.TrimPrefix(raw, "http://")
		insecure = true
	case strings.HasPrefix(raw, "https://"):
		raw = strings.TrimPrefix(raw, "https://")
	}
	return strings.TrimRight(raw, "/"), insecure
}

func parseBool(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
