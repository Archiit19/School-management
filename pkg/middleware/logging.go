package middleware

import (
	"net/http"
	"time"

	"github.com/Archiit19/School-management/pkg/correlation"
	"github.com/Archiit19/School-management/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const requestIDKey = "request_id"

// NewEngine returns a Gin engine with structured request logging and panic recovery.
// Prefer this over gin.Default() so HTTP access logs use pkg/logger.
func NewEngine() *gin.Engine {
	r := gin.New()
	r.Use(RequestLogger(), Recovery())
	return r
}

// RequestLogger emits one structured log line per HTTP request.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		reqID := c.GetHeader(correlation.RequestIDHeader)
		if reqID == "" {
			reqID = uuid.New().String()
		}
		c.Set(requestIDKey, reqID)
		c.Header(correlation.RequestIDHeader, reqID)
		c.Request = c.Request.WithContext(correlation.ContextWithRequestID(c.Request.Context(), reqID))

		c.Next()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		pairs := []any{
			"request_id", reqID,
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		}

		if ua := c.Request.UserAgent(); ua != "" {
			pairs = append(pairs, "user_agent", ua)
		}
		if uid, ok := c.Get("user_id"); ok {
			pairs = append(pairs, "user_id", uid)
		}
		if role, ok := c.Get("role_name"); ok {
			pairs = append(pairs, "role_name", role)
		}

		fields := logger.KV(pairs...)
		status := c.Writer.Status()
		switch {
		case status >= 500:
			logger.Error("request", fields...)
		case status >= 400:
			logger.Warn("request", fields...)
		default:
			logger.Info("request", fields...)
		}
	}
}

// GetRequestID returns the request ID for the current Gin context.
func GetRequestID(c *gin.Context) string {
	if v, ok := c.Get(requestIDKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// Recovery logs panics with request context and returns 500.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				reqID, _ := c.Get(requestIDKey)
				logger.Error("panic recovered", logger.KV(
					"request_id", reqID,
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"panic", recovered,
				)...)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
			}
		}()
		c.Next()
	}
}
