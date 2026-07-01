package config

import (
	"fmt"
	"strings"
)

const (
	defaultJWTSecret = "super-secret-jwt-key-change-in-production"
	defaultDevToken  = "dev-internal-token-change-in-production"
)

// IsProduction reports whether APP_ENV is set to production.
func IsProduction() bool {
	return strings.EqualFold(strings.TrimSpace(GetEnv("APP_ENV", "development")), "production")
}

// ValidateCommon rejects insecure defaults when APP_ENV=production.
func ValidateCommon(jwtSecret, internalToken string) error {
	if !IsProduction() {
		return nil
	}

	jwtSecret = strings.TrimSpace(jwtSecret)
	if jwtSecret == "" || jwtSecret == defaultJWTSecret {
		return fmt.Errorf("JWT_SECRET must be set to a strong value when APP_ENV=production")
	}

	internalToken = strings.TrimSpace(internalToken)
	if internalToken == "" || internalToken == defaultDevToken {
		return fmt.Errorf("INTERNAL_SERVICE_TOKEN must be set to a strong value when APP_ENV=production")
	}
	return nil
}
