package config

import (
	"strings"
	"testing"
)

func TestPostgresDSNIncludesSSLMode(t *testing.T) {
	dsn := Postgres{
		Host:     "db.example.com",
		Port:     "5432",
		User:     "app",
		Password: "secret",
		Name:     "school",
		SSLMode:  "require",
	}.DSN()

	if !strings.Contains(dsn, "sslmode=require") {
		t.Fatalf("DSN %q missing sslmode=require", dsn)
	}
}

func TestValidateCommonSkipsOutsideProduction(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	if err := ValidateCommon("", ""); err != nil {
		t.Fatalf("expected nil outside production, got %v", err)
	}
}

func TestValidateCommonRejectsWeakSecretsInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	if err := ValidateCommon("super-secret-jwt-key-change-in-production", "token"); err == nil {
		t.Fatal("expected JWT validation error")
	}
	if err := ValidateCommon("strong-jwt-secret", ""); err == nil {
		t.Fatal("expected internal token validation error")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexSubstring(s, sub) >= 0)
}

func indexSubstring(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
