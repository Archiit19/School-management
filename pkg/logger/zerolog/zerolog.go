package zerolog

import (
	"fmt"

	"github.com/Archiit19/School-management/pkg/logger/core"
)

func init() {
	core.RegisterBackend(core.BackendZerolog, New)
}

// New builds a zerolog-backed core.Logger.
// Add github.com/rs/zerolog to pkg/go.mod and implement the adapter to enable this backend.
func New(cfg core.Config) (core.Logger, error) {
	_ = cfg
	return nil, fmt.Errorf("logger backend %q is not implemented yet; set LOG_BACKEND=slog or implement pkg/logger/zerolog", core.BackendZerolog)
}
