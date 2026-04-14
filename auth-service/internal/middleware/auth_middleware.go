package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTAuth validates the Bearer token and injects claims into context.
func JWTAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		tokenStr := parts[1]

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		userIDStr, _ := claims["user_id"].(string)
		schoolIDStr, _ := claims["school_id"].(string)
		roleIDStr, _ := claims["role_id"].(string)
		roleName, _ := claims["role_name"].(string)
		email, _ := claims["email"].(string)

		userID, _ := uuid.Parse(userIDStr)
		schoolID, _ := uuid.Parse(schoolIDStr)
		roleID, _ := uuid.Parse(roleIDStr)

		c.Set("user_id", userID)
		c.Set("school_id", schoolID)
		c.Set("role_id", roleID)
		c.Set("role_name", roleName)
		c.Set("email", email)
		c.Set("permissions", parsePermissions(claims))

		c.Next()
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

// RequirePermission checks that the authenticated user's role has the required permission.
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

func parsePermissions(claims jwt.MapClaims) []string {
	raw, ok := claims["permissions"]
	if !ok || raw == nil {
		return nil
	}

	arr, ok := raw.([]interface{})
	if !ok {
		return nil
	}

	perms := make([]string, 0, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok {
			perms = append(perms, s)
		}
	}
	return perms
}
