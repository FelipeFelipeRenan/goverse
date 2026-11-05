package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("GOVERSE_DATABASE_URL", "postgres://x")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.Role != RoleAll || c.HTTPAddr != ":8080" || c.ShutdownTimeout.Seconds() != 10 {
		t.Fatalf("defaults inesperados: %+v", c)
	}
}

func TestLoadValidation(t *testing.T) {
	t.Run("sem banco", func(t *testing.T) {
		t.Setenv("GOVERSE_DATABASE_URL", "")
		if _, err := Load(); err == nil {
			t.Fatal("esperava erro")
		}
	})
	t.Run("role inválida", func(t *testing.T) {
		t.Setenv("GOVERSE_DATABASE_URL", "postgres://x")
		t.Setenv("GOVERSE_ROLE", "foo")
		if _, err := Load(); err == nil {
			t.Fatal("esperava erro")
		}
	})
	t.Run("timeout inválido", func(t *testing.T) {
		t.Setenv("GOVERSE_DATABASE_URL", "postgres://x")
		t.Setenv("GOVERSE_SHUTDOWN_TIMEOUT", "abc")
		if _, err := Load(); err == nil {
			t.Fatal("esperava erro")
		}
	})
}
