// Package config carrega a configuração do processo a partir de variáveis de ambiente.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
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

	JWTPrivateKeyPath string   // chave RSA que assina as sessões (a pública é derivada)
	CORSOrigins       []string // origens do frontend (cookies exigem lista explícita)
	AutoMigrate       bool     // aplica as migrations embutidas no boot
	TrustProxy        bool     // atrás de proxy/ingress confiável: usa X-Forwarded-For para o IP do cliente

	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
}

// IsProd indica cookies Secure + SameSite=None (frontend cross-site).
func (c Config) IsProd() bool { return c.Env == "prod" }

// Load lê as variáveis GOVERSE_* e valida o resultado.
func Load() (Config, error) {
	c := Config{
		ServiceName: get("GOVERSE_SERVICE_NAME", "goverse"),
		Env:         get("GOVERSE_ENV", "dev"),
		Role:        strings.ToLower(get("GOVERSE_ROLE", RoleAll)),
		HTTPAddr:    get("GOVERSE_HTTP_ADDR", ":8080"),
		DatabaseURL: os.Getenv("GOVERSE_DATABASE_URL"),
		LogLevel:    get("GOVERSE_LOG_LEVEL", "info"),

		JWTPrivateKeyPath: os.Getenv("GOVERSE_JWT_PRIVATE_KEY_PATH"),
		CORSOrigins:       splitList(os.Getenv("GOVERSE_CORS_ORIGINS")),

		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  os.Getenv("GOOGLE_REDIRECT_URI"),
	}

	auto, err := strconv.ParseBool(get("GOVERSE_AUTO_MIGRATE", "true"))
	if err != nil {
		return Config{}, fmt.Errorf("GOVERSE_AUTO_MIGRATE inválido: %w", err)
	}
	c.AutoMigrate = auto

	trust, err := strconv.ParseBool(get("GOVERSE_TRUST_PROXY", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("GOVERSE_TRUST_PROXY inválido: %w", err)
	}
	c.TrustProxy = trust

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
	if c.JWTPrivateKeyPath == "" {
		return Config{}, errors.New("GOVERSE_JWT_PRIVATE_KEY_PATH é obrigatório")
	}
	return c, nil
}

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
