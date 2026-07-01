package config

// Common holds settings shared by most microservices.
type Common struct {
	JWTSecret            string
	Port                 string
	InternalServiceToken string
}

// LoadCommonFromEnv reads cross-service settings from the environment.
func LoadCommonFromEnv() Common {
	return Common{
		JWTSecret:            GetEnv("JWT_SECRET", "super-secret-jwt-key-change-in-production"),
		Port:                 GetEnv("PORT", "8080"),
		InternalServiceToken: GetEnv("INTERNAL_SERVICE_TOKEN", ""),
	}
}
