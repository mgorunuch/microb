package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/mgorunuch/microb/app/core"
	"github.com/mgorunuch/microb/app/core/postgres"
)

func main() {
	// Define flags
	reason := flag.String("reason", "", "Reason for storing the domains")
	flag.Parse()

	if *reason == "" {
		fmt.Println("Error: -reason flag is required")
		os.Exit(1)
	}

	// Initialize core and postgres
	core.Init()
	cleanup := postgres.Init(context.Background())
	defer cleanup()

	ctx := context.Background()
	reasons := []string{*reason}

	core.ProcessLines(core.SimpleConfig[string]{
		RunFunc: func(domain string) (string, error) {
			_, err := postgres.CreateDomain(ctx, domain, reasons)
			if err != nil {
				core.Logger.Errorf("Failed to store domain %s: %v", domain, err)
				return domain, err
			}
			core.Logger.Infof("Stored domain: %s", domain)
			return domain, nil
		},
	})
}
