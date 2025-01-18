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

func main() {
	// Define flags
	reason := flag.String("reason", "", "Reason for storing the links")
	visitID := flag.String("visit", "", "ID of the chrome visit to associate URLs with")

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
	flags := []string{*reason}

	core.ProcessLines(core.SimpleConfig[string]{
		ThreadsCount: 10,
		KeyFunc:      nil,
		RunFunc: func(url string) (string, error) {
			// Clean the URL by removing any control characters
			cleanURL := strings.TrimSpace(url)
			if cleanURL == "" {
				return url, fmt.Errorf("empty URL after cleaning")
			}

			// Store the URL
			storedURL, err := postgres.CreateURL(ctx, cleanURL, flags)
			if err != nil {
				core.Logger.Errorf("Failed to store URL %s: %v", cleanURL, err)
				return url, err
			}

			// Get the specified visit with all its URLs
			existingVisit := &postgres.ChromeVisit{}
			var primaryURLID string
			err = postgres.Pool.QueryRow(ctx, `
				select cv.id, cv.opened_at, cv.success, cv.reason, cv.title, cv.url_id
				from chrome_visits cv
				where cv.id = $1
			`, *visitID).Scan(&existingVisit.ID, &existingVisit.OpenedAt, &existingVisit.Success, &existingVisit.Reason, &existingVisit.Title, &primaryURLID)

			if err != nil {
				core.Logger.Errorf("Failed to get chrome visit %s: %v", *visitID, err)
				return url, err
			}

			// Get existing URLs for this visit
			rows, err := postgres.Pool.Query(ctx, `
				select u.id, u.raw, u.flags, u.hostname, u.path, u.scheme, u.query, u.fragment, u.created_at
				from urls u
				join url_visits uv on uv.url_id = u.id
				where uv.visit_id = $1
			`, existingVisit.ID)
			if err != nil {
				core.Logger.Errorf("Failed to get URLs for visit %s: %v", *visitID, err)
				return url, err
			}
			defer rows.Close()

			var urls []*postgres.URL
			for rows.Next() {
				var url postgres.URL
				err = rows.Scan(
					&url.ID,
					&url.Raw,
					&url.Flags,
					&url.Hostname,
					&url.Path,
					&url.Scheme,
					&url.Query,
					&url.Fragment,
					&url.CreatedAt,
				)
				if err != nil {
					core.Logger.Errorf("Failed to scan URL: %v", err)
					continue
				}
				urlCopy := url
				urls = append(urls, &urlCopy)
				if url.ID == primaryURLID {
					existingVisit.URL = &urlCopy
				}
			}

			existingVisit.URLs = urls

			// Add the new URL if not already present
			found := false
			for _, u := range existingVisit.URLs {
				if u.Raw == cleanURL {
					found = true
					break
				}
			}
			if !found {
				existingVisit.URLs = append(existingVisit.URLs, storedURL)
			}

			// Insert new URL-visit relationship rather than recreating the whole visit
			_, err = postgres.Pool.Exec(ctx, `
				insert into url_visits (url_id, visit_id)
				values ($1, $2)
				on conflict do nothing
			`, storedURL.ID, existingVisit.ID)
			if err != nil {
				core.Logger.Errorf("Failed to create URL-visit relationship: %v", err)
				return url, err
			}

			core.Logger.Infof("Added URL %s to visit %s", cleanURL, *visitID)

			return url, nil
		},
		OutputFunc: nil,
		SleepTime:  0,
		Unique:     false,
	})
}
