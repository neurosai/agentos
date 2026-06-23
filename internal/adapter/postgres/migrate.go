package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// Migrate runs goose migrations against the pool DSN.
func Migrate(ctx context.Context, databaseURL string) error {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open db for migrate: %w", err)
	}
	defer db.Close()
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	dir, err := migrationsDir()
	if err != nil {
		return err
	}
	return goose.UpContext(ctx, db, dir)
}

func migrationsDir() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("caller info")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../../migrations")), nil
}
