package config

import (
	"reflect"
	"testing"
)

func baseEnv(t *testing.T) {
	t.Helper()
	t.Setenv("GOVERSE_DATABASE_URL", "postgres://x")
	t.Setenv("GOVERSE_JWT_PRIVATE_KEY_PATH", "/keys/private.pem")
}

func TestLoadDefaults(t *testing.T) {
	baseEnv(t)
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.Role != RoleAll || c.HTTPAddr != ":8080" || c.ShutdownTimeout.Seconds() != 10 || !c.AutoMigrate || c.IsProd() {
		t.Fatalf("defaults inesperados: %+v", c)
	}
}

func TestLoadParsing(t *testing.T) {
	baseEnv(t)
	t.Setenv("GOVERSE_ENV", "prod")
	t.Setenv("GOVERSE_AUTO_MIGRATE", "false")
	t.Setenv("GOVERSE_CORS_ORIGINS", " http://a.com, ,https://b.com ")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !c.IsProd() || c.AutoMigrate {
		t.Fatalf("env/auto-migrate: %+v", c)
	}
	if want := []string{"http://a.com", "https://b.com"}; !reflect.DeepEqual(c.CORSOrigins, want) {
		t.Fatalf("cors = %v, quero %v", c.CORSOrigins, want)
	}
}

func TestLoadValidation(t *testing.T) {
	cases := map[string]map[string]string{
		"sem banco":        {"GOVERSE_DATABASE_URL": ""},
		"sem chave jwt":    {"GOVERSE_JWT_PRIVATE_KEY_PATH": ""},
		"role inválida":    {"GOVERSE_ROLE": "foo"},
		"timeout inválido": {"GOVERSE_SHUTDOWN_TIMEOUT": "abc"},
		"migrate inválido": {"GOVERSE_AUTO_MIGRATE": "talvez"},
	}
	for name, env := range cases {
		t.Run(name, func(t *testing.T) {
			baseEnv(t)
			for k, v := range env {
				t.Setenv(k, v)
			}
			if _, err := Load(); err == nil {
				t.Fatal("esperava erro")
			}
		})
	}
}
