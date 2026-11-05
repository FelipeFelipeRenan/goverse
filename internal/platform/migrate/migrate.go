// Package migrate aplica migrations SQL embutidas (somente "up"), uma por arquivo, em ordem
// lexicográfica. Cada módulo tem o próprio schema Postgres e a própria tabela de controle.
// Substitui o binário golang-migrate que era baixado via curl nos Dockerfiles.
package migrate

import (
	"context"
	"fmt"
	"io/fs"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// lockKey serializa migrations entre instâncias (advisory lock da sessão).
const lockKey int64 = 7_274_103

func Run(ctx context.Context, pool *pgxpool.Pool, schema string, migrations fs.FS) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", lockKey); err != nil {
		return fmt.Errorf("lock de migration: %w", err)
	}
	defer func() { _, _ = conn.Exec(context.WithoutCancel(ctx), "SELECT pg_advisory_unlock($1)", lockKey) }()

	s := pgx.Identifier{schema}.Sanitize()
	if _, err := conn.Exec(ctx, "CREATE SCHEMA IF NOT EXISTS "+s); err != nil {
		return err
	}
	if _, err := conn.Exec(ctx, "CREATE TABLE IF NOT EXISTS "+s+".schema_migrations ("+
		"version text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())"); err != nil {
		return err
	}

	files, err := fs.Glob(migrations, "*.sql")
	if err != nil {
		return err
	}
	sort.Strings(files)

	for _, name := range files {
		var done bool
		if err := conn.QueryRow(ctx,
			"SELECT EXISTS (SELECT 1 FROM "+s+".schema_migrations WHERE version = $1)", name).Scan(&done); err != nil {
			return err
		}
		if done {
			continue
		}
		sqlBytes, err := fs.ReadFile(migrations, name)
		if err != nil {
			return err
		}
		tx, err := conn.Begin(ctx)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("migration %s/%s: %w", schema, name, err)
		}
		if _, err := tx.Exec(ctx, "INSERT INTO "+s+".schema_migrations (version) VALUES ($1)", name); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
	}
	return nil
}
