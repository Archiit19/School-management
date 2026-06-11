package core

import "fmt"

// Factory constructs a Logger for a given configuration.
type Factory func(Config) (Logger, error)

var registry = map[Backend]Factory{}

// RegisterBackend adds a logging backend factory. Typically called from init() in backend packages.
func RegisterBackend(backend Backend, factory Factory) {
	registry[backend] = factory
}

// New constructs a Logger using the registered backend factory.
func New(cfg Config) (Logger, error) {
	if cfg.Output == nil {
		return nil, fmt.Errorf("logger: output writer is required")
	}
	cfg.Level = ParseLevel(cfg.Level)

	backend := cfg.Backend
	if backend == "" {
		backend = BackendSlog
	}

	factory, ok := registry[backend]
	if !ok {
		return nil, fmt.Errorf("logger: backend %q is not registered", backend)
	}
	return factory(cfg)
}
