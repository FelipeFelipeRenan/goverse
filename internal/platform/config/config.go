package config

import (
	"fmt"
	"os"
)

// AppConfig contains the process-wide configuration for the modular monolith.
type AppConfig struct {
	Environment       string
	HTTPAddress       string
	DatabaseURL       string
	JWTPrivateKeyPath string
}

// Load returns the configuration shared by all internal modules.
func Load() AppConfig {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	return AppConfig{
		Environment:       valueOrDefault("ENV", "development"),
		HTTPAddress:       ":" + port,
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		JWTPrivateKeyPath: os.Getenv("JWT_PRIVATE_KEY_PATH"),
	}
}

// RequireEnv validates that every required environment variable is present.
func RequireEnv(vars ...string) error {
	var missing []string

	for _, variable := range vars {
		if os.Getenv(variable) == "" {
			missing = append(missing, variable)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("variaveis de ambiente obrigatórias ausentes: %v", missing)
	}

	return nil
}

func valueOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
