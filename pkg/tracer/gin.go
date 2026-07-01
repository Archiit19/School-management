package tracer

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// GinMiddleware returns Gin middleware that creates a server span per request.
// When tracing is disabled, returns a no-op handler.
func GinMiddleware(service string) gin.HandlerFunc {
	return GinMiddlewareWithEnabled(service, Enabled())
}

// GinMiddlewareWithEnabled returns Gin middleware controlled by an explicit enabled flag.
func GinMiddlewareWithEnabled(service string, enabled bool) gin.HandlerFunc {
	if !enabled {
		return func(c *gin.Context) { c.Next() }
	}
	return otelgin.Middleware(service, otelgin.WithFilter(skipNoiseRoutes))
}

func skipNoiseRoutes(r *http.Request) bool {
	path := r.URL.Path
	if path == "/health" || path == "/health/live" || path == "/health/ready" {
		return false
	}
	if strings.HasPrefix(path, "/swagger") {
		return false
	}
	return true
}
