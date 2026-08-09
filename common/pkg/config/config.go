// Package config provides compatibility during the migration to the modular monolith.
package config

import platformconfig "github.com/FelipeFelipeRenan/goverse/internal/platform/config"

// RequireEnv delegates environment validation to the centralized configuration package.
func RequireEnv(vars ...string) error {
	return platformconfig.RequireEnv(vars...)
}
