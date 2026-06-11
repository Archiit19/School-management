// Package backends registers all built-in logger implementations.
// Import this package (or pkg/logger, which imports it) to enable backend selection.
package backends

import (
	_ "github.com/Archiit19/School-management/pkg/logger/slog"
	_ "github.com/Archiit19/School-management/pkg/logger/zap"
	_ "github.com/Archiit19/School-management/pkg/logger/zerolog"
)
