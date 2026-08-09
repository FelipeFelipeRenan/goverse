package config

import "testing"

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("APP_PORT", "")
	t.Setenv("ENV", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("JWT_PRIVATE_KEY_PATH", "")

	got := Load()

	if got.Environment != "development" {
		t.Fatalf("Environment = %q, want %q", got.Environment, "development")
	}
	if got.HTTPAddress != ":8080" {
		t.Fatalf("HTTPAddress = %q, want %q", got.HTTPAddress, ":8080")
	}
	if got.DatabaseURL != "" {
		t.Fatalf("DatabaseURL = %q, want empty value", got.DatabaseURL)
	}
	if got.JWTPrivateKeyPath != "" {
		t.Fatalf("JWTPrivateKeyPath = %q, want empty value", got.JWTPrivateKeyPath)
	}
}

func TestRequireEnv(t *testing.T) {
	t.Setenv("MONOLITH_CONFIG_TEST", "set")

	if err := RequireEnv("MONOLITH_CONFIG_TEST"); err != nil {
		t.Fatalf("RequireEnv() error = %v, want nil", err)
	}
	if err := RequireEnv("MONOLITH_CONFIG_MISSING"); err == nil {
		t.Fatal("RequireEnv() error = nil, want missing-variable error")
	}
}
