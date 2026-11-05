// Command goverse é o binário único da plataforma (monolito modular).
// O papel (GOVERSE_ROLE) permite, no futuro, separar o tráfego realtime sem mudar o código.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FelipeFelipeRenan/goverse/internal/identity"
	"github.com/FelipeFelipeRenan/goverse/internal/platform/authn"
	"github.com/FelipeFelipeRenan/goverse/internal/platform/config"
	"github.com/FelipeFelipeRenan/goverse/internal/platform/database"
	"github.com/FelipeFelipeRenan/goverse/internal/platform/httpx"
	"github.com/FelipeFelipeRenan/goverse/internal/platform/module"
	"github.com/FelipeFelipeRenan/goverse/internal/platform/telemetry"
	"github.com/FelipeFelipeRenan/goverse/internal/rooms"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "erro fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	slog.SetDefault(telemetry.NewLogger(cfg.LogLevel, cfg.ServiceName, cfg.Env))

	shutdownTelemetry, err := telemetry.Setup(ctx, cfg.ServiceName, cfg.Env)
	if err != nil {
		return fmt.Errorf("telemetria: %w", err)
	}

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("banco: %w", err)
	}
	defer pool.Close()

	keys, err := authn.LoadKeys(cfg.JWTPrivateKeyPath)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		pctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(pctx); err != nil {
			http.Error(w, "db indisponível", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	identityMod := identity.New(pool, keys, identity.Config{
		Secure:             cfg.IsProd(),
		GoogleClientID:     cfg.GoogleClientID,
		GoogleClientSecret: cfg.GoogleClientSecret,
		GoogleRedirectURL:  cfg.GoogleRedirectURL,
	})
	roomsMod := rooms.New(pool, keys, identityMod.Directory())
	// chat entra no próximo lote e receberá roomsMod.Membership().

	// Ordem importa: rooms tem FK para identity.users, então identity migra primeiro.
	modules := []module.Module{identityMod, roomsMod}

	for _, m := range modules {
		if mg, ok := m.(module.Migrator); ok && cfg.AutoMigrate {
			if err := mg.Migrate(ctx); err != nil {
				return fmt.Errorf("migrations de %s: %w", m.Name(), err)
			}
		}
		slog.Info("registrando módulo", "module", m.Name(), "role", cfg.Role)
		m.Register(mux)
	}

	// Sem WriteTimeout global: conexões WebSocket são longas. Timeouts por rota virão com os módulos.
	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpx.Wrap(httpx.CORS(cfg.CORSOrigins, mux)),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("goverse iniciado", "addr", cfg.HTTPAddr, "role", cfg.Role)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		slog.Info("encerrando")
	}

	sctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(sctx); err != nil {
		slog.Error("shutdown http", "err", err)
	}
	return shutdownTelemetry(sctx)
}
