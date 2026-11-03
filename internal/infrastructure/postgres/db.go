package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dealngon/backend/internal/domain"
	"github.com/dealngon/backend/migrations"
	"github.com/golang-migrate/migrate/v4"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	var last error
	for attempt := 0; attempt < 30; attempt++ {
		pool, err := pgxpool.New(ctx, dsn)
		if err == nil {
			pingErr := pool.Ping(ctx)
			if pingErr == nil {
				return pool, nil
			}
			pool.Close()
			last = pingErr
		} else {
			last = err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return nil, fmt.Errorf("connect database: %w", last)
}

func Migrate(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	driver, err := pgxmigrate.WithInstance(db, &pgxmigrate.Config{})
	if err != nil {
		return err
	}
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}
	migrator, err := migrate.NewWithInstance("iofs", source, "pgx", driver)
	if err != nil {
		return err
	}
	err = migrator.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505", "23503":
			return domain.NewError("CONFLICT", "Dữ liệu bị trùng hoặc đang được sử dụng", domain.ErrConflict)
		}
	}
	return err
}
