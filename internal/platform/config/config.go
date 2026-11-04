// Package config carrega a configuração do processo a partir de variáveis de ambiente.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// Roles possíveis do binário. "all" roda tudo no mesmo processo.
const (
	RoleAll      = "all"
	RoleAPI      = "api"
	RoleRealtime = "realtime"
)

type Config struct {
	ServiceName     string
	Env             string
	Role            string
	HTTPAddr        string
	DatabaseURL     string
	LogLevel        string
	ShutdownTimeout time.Duration
}

// Load lê as variáveis GOVERSE_* e valida o resultado.
func Load() (Config, error) {
	c := Config{
		ServiceName: get("GOVERSE_SERVICE_NAME", "goverse"),
		Env:         get("GOVERSE_ENV", "dev"),
		Role:        strings.ToLower(get("GOVERSE_ROLE", RoleAll)),
		HTTPAddr:    get("GOVERSE_HTTP_ADDR", ":8080"),
		DatabaseURL: os.Getenv("GOVERSE_DATABASE_URL"),
		LogLevel:    get("GOVERSE_LOG_LEVEL", "info"),
	}

	d, err := time.ParseDuration(get("GOVERSE_SHUTDOWN_TIMEOUT", "10s"))
	if err != nil {
		return Config{}, fmt.Errorf("GOVERSE_SHUTDOWN_TIMEOUT inválido: %w", err)
	}
	c.ShutdownTimeout = d

	switch c.Role {
	case RoleAll, RoleAPI, RoleRealtime:
	default:
		return Config{}, fmt.Errorf("GOVERSE_ROLE inválido: %q (use all, api ou realtime)", c.Role)
	}
	if c.DatabaseURL == "" {
		return Config{}, errors.New("GOVERSE_DATABASE_URL é obrigatório")
	}
	return c, nil
}

func get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
