package main

import (
	"context"
	"os"

	"github.com/mgorunuch/microb/app/core"
	"github.com/mgorunuch/microb/app/core/postgres"
)

func main() {
	core.Init()
	cleanup := postgres.Init(context.Background())
	defer cleanup()

	ctx := context.Background()

	if err := postgres.Migrate(ctx); err != nil {
		core.Logger.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	core.Logger.Info("Successfully ran all migrations")
}
