package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mgorunuch/microb/app/core"
	"github.com/mgorunuch/microb/app/core/postgres"
)

// Define flags
var reason = flag.String("reason", "", "Reason for storing the links")
var visitID = flag.String("visit", "", "ID of the chrome visit to associate URLs with")

func process(ctx context.Context, url string) (string, error) {
	// Clean the URL by removing any control characters
	cleanURL := strings.TrimSpace(url)
	if cleanURL == "" {
		return url, fmt.Errorf("empty URL after cleaning")
	}

	// Get all urls
	urlModel, err := postgres.UrlRepo.UpsertByRaw(ctx, cleanURL)
	if err != nil {
		core.Logger.Errorf("Failed to upsert URL %s: %v", cleanURL, err)
		return url, err
	}

	err = postgres.URLVisitRepo.UpsertRaw(ctx, urlModel.Id, *visitID)
	if err != nil {
		return url, err
	}

	return url, nil
}

func main() {
	flag.Parse()

	if *reason == "" {
		fmt.Println("Error: -reason flag is required")
		os.Exit(1)
	}

	ctx := context.Background()

	// Initialize core and postgres
	core.Init()
	defer postgres.Init(ctx)()

	core.ProcessLines(core.SimpleConfig[string]{
		Ctx:          ctx,
		ThreadsCount: 10,
		RunFunc:      process,
	})
}
