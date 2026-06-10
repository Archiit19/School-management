package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RequirePermission checks that the user's role has the required permission.
// super_admin bypasses all permission checks.
func RequirePermission(required string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleName, _ := c.Get("role_name")
		if roleName == "super_admin" {
			c.Next()
			return
		}

		perms, exists := c.Get("permissions")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "no permissions found"})
			return
		}

		permList, ok := perms.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid permissions data"})
			return
		}

		for _, p := range permList {
			if p == required {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "permission denied, requires: " + required,
		})
	}
}

// RequireAnyPermission allows access if the user has at least one listed permission.
func RequireAnyPermission(required ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleName, _ := c.Get("role_name")
		if roleName == "super_admin" {
			c.Next()
			return
		}

		perms, exists := c.Get("permissions")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "no permissions found"})
			return
		}

		permList, ok := perms.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid permissions data"})
			return
		}

		for _, required := range required {
			for _, p := range permList {
				if p == required {
					c.Next()
					return
				}
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "permission denied",
		})
	}
}

// RequireRole checks that the authenticated user has one of the allowed roles.
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleName, exists := c.Get("role_name")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "role information missing"})
			return
		}

		role, ok := roleName.(string)
		if !ok || role == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid role"})
			return
		}

		for _, allowed := range allowedRoles {
			if role == allowed {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "insufficient permissions, required roles: " + strings.Join(allowedRoles, ", "),
		})
	}
}

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
