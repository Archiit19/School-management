package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const InternalTokenHeader = "X-Internal-Token"

// RequireInternalToken validates X-Internal-Token against expectedSecret.
// If expectedSecret is empty, internal APIs are disabled (503).
func RequireInternalToken(expectedSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.TrimSpace(expectedSecret) == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "internal service API is not configured",
			})
			return
		}

		got := strings.TrimSpace(c.GetHeader(InternalTokenHeader))
		if got == "" || got != expectedSecret {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid internal token"})
			return
		}

		c.Next()
	}
}
