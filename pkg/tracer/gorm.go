package tracer

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

// InstrumentGORM registers the OpenTelemetry plugin on db when tracing is enabled.
func InstrumentGORM(db *gorm.DB) error {
	if db == nil || !Enabled() {
		return nil
	}
	if err := db.Use(tracing.NewPlugin(
		tracing.WithDBSystem("postgresql"),
		tracing.WithTracerProvider(Provider()),
	)); err != nil {
		return fmt.Errorf("tracer: gorm plugin: %w", err)
	}
	return nil
}
