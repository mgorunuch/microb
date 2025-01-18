package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mgorunuch/microb/app/core"

	sq "github.com/Masterminds/squirrel"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func POSTGRES_HOST() string {
	return core.Env.GetDefault("POSTGRES_HOST", "localhost")
}

func POSTGRES_PORT() string {
	return core.Env.GetDefault("POSTGRES_PORT", "5432")
}

func POSTGRES_DB() string {
	return core.Env.GetDefault("POSTGRES_DB", "postgres")
}

func POSTGRES_USER() string {
	return core.Env.GetDefault("POSTGRES_USER", "postgres")
}

func POSTGRES_PASSWORD() string {
	return core.Env.Get("POSTGRES_PASSWORD", true)
}

var Pool *pgxpool.Pool

func Init(ctx context.Context) func() error {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		POSTGRES_USER(),
		POSTGRES_PASSWORD(),
		POSTGRES_HOST(),
		POSTGRES_PORT(),
		POSTGRES_DB(),
	)

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		core.Logger.Fatalf("Unable to parse connection string: %v\n", err)
	}

	Pool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		core.Logger.Fatalf("Unable to create connection pool: %v\n", err)
	}

	// Verify connection
	if err := Pool.Ping(ctx); err != nil {
		core.Logger.Fatalf("Unable to connect to database: %v\n", err)
	}

	// Run migrations
	if err := Migrate(ctx); err != nil {
		core.Logger.Fatalf("Failed to run migrations: %v\n", err)
	}

	return func() error {
		Pool.Close()
		return nil
	}
}
