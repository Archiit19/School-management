package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pkgconfig "github.com/Archiit19/School-management/pkg/config"
	"github.com/Archiit19/School-management/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Config controls HTTP server timeouts and graceful shutdown.
type Config struct {
	Addr            string
	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
}

// DefaultConfig returns production-oriented server defaults for port.
func DefaultConfig(port string) Config {
	return Config{
		Addr:            fmt.Sprintf(":%s", port),
		ShutdownTimeout: 15 * time.Second,
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     60 * time.Second,
	}
}

// LoadConfigFromEnv builds server settings using standard env vars.
func LoadConfigFromEnv(port string) Config {
	def := DefaultConfig(port)
	return Config{
		Addr:            def.Addr,
		ShutdownTimeout: pkgconfig.GetEnvDuration("HTTP_SHUTDOWN_TIMEOUT", def.ShutdownTimeout),
		ReadTimeout:     pkgconfig.GetEnvDuration("HTTP_READ_TIMEOUT", def.ReadTimeout),
		WriteTimeout:    pkgconfig.GetEnvDuration("HTTP_WRITE_TIMEOUT", def.WriteTimeout),
		IdleTimeout:     pkgconfig.GetEnvDuration("HTTP_IDLE_TIMEOUT", def.IdleTimeout),
	}
}

// Run starts the Gin engine and shuts down gracefully on SIGINT/SIGTERM.
func Run(engine *gin.Engine, cfg Config) error {
	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      engine,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("http server listening", logger.AddField("addr", cfg.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-quit:
		logger.Info("shutdown signal received", logger.AddField("signal", sig.String()))
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}
	_ = logger.L().Sync()
	logger.Info("http server stopped")
	return nil
}
