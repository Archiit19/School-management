package config

// Common holds settings shared by most microservices.
type Common struct {
	JWTSecret            string
	Port                 string
	InternalServiceToken string
}
