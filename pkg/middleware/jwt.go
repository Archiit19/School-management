package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const InternalTokenHeader = "X-Internal-Token"

// JWTOption configures optional JWT claim extraction.
type JWTOption func(*jwtSettings)

type jwtSettings struct {
	uuidClaims   []string
	stringClaims []string
}

// WithUUIDClaim parses claim as UUID and stores it in gin context under the same key.
func WithUUIDClaim(claim string) JWTOption {
	return func(s *jwtSettings) {
		s.uuidClaims = append(s.uuidClaims, claim)
	}
}

// WithStringClaim copies a string claim into gin context under the same key.
func WithStringClaim(claim string) JWTOption {
	return func(s *jwtSettings) {
		s.stringClaims = append(s.stringClaims, claim)
	}
}

// JWTAuth validates Bearer JWT and injects standard claims into context.
func JWTAuth(jwtSecret string, opts ...JWTOption) gin.HandlerFunc {
	settings := jwtSettings{}
	for _, opt := range opts {
		opt(&settings)
	}

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

		token, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
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
		roleName, _ := claims["role_name"].(string)

		userID, _ := uuid.Parse(userIDStr)
		schoolID, _ := uuid.Parse(schoolIDStr)

		c.Set("user_id", userID)
		c.Set("school_id", schoolID)
		c.Set("role_name", roleName)
		c.Set("permissions", parsePermissions(claims))

		if sidStr, ok := claims["student_id"].(string); ok && sidStr != "" {
			if sid, err := uuid.Parse(sidStr); err == nil {
				c.Set("student_id", sid)
			}
		}

		for _, claim := range settings.uuidClaims {
			if val, ok := claims[claim].(string); ok && val != "" {
				if id, err := uuid.Parse(val); err == nil {
					c.Set(claim, id)
				}
			}
		}

		for _, claim := range settings.stringClaims {
			if val, ok := claims[claim].(string); ok {
				c.Set(claim, val)
			}
		}

		c.Next()
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
