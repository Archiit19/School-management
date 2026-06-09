package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireInternalToken(expected string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if expected == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "internal API not configured"})
			return
		}
		if c.GetHeader("X-Internal-Token") != expected {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid internal token"})
			return
		}
		c.Next()
	}
}
