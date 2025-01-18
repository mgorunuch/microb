package main

import (
	"context"
	"flag"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/jackc/pgx/v5"
	"github.com/mgorunuch/microb/app/core"
	"github.com/mgorunuch/microb/app/core/chrome"
	"github.com/mgorunuch/microb/app/core/postgres"
)

var reasonFlag = flag.String("reason", "manual check", "Reason for visiting URLs")

func main() {
	ctx := context.Background()

	core.Init()

	cleanup := postgres.Init(ctx)
	defer cleanup()

	chromeController := chrome.NewWindow()
	defer chromeController.Close()

	chromeCtx := chromeController.NewContext()
	defer chromeCtx.Close()

	// Create control server
	server, err := NewControlServer(chromeCtx)
	if err != nil {
		core.Logger.Fatal(err)
	}

	// Start HTTP server
	var wg sync.WaitGroup

	go server.ListenAndServe(&wg)()

	core.ReadAllLines(func(rawUrl string) {
		url := strings.TrimSpace(rawUrl)

		urlModel, err := postgres.UrlRepo.UpsertByRaw(ctx, url)
		if err != nil {
			core.Logger.Errorf("Failed to upsert URL %s: %v", url, err)
			return
		}

		lastVisit, err := postgres.ChromeVisitRepo.GetLastVisitAfter(ctx, urlModel.Id, time.Hour*24*365)
		if err != nil && err != pgx.ErrNoRows {
			core.Logger.Errorf("Failed to check last visit: %v", err)
			return
		}

		if lastVisit != nil {
			core.Logger.Infof("Skipping URL %s - already visited on %s", url, lastVisit.OpenedAt)
			return
		}

		visit := &postgres.ChromeVisitModel{
			UrlId:     urlModel.Id,
			OpenedAt:  time.Now(),
			Success:   true,
			Reason:    *reasonFlag,
			CreatedAt: time.Now(),
		}

		// Add timeout - 5 minutes for review
		tabCtx := chromeCtx.NewContextWithTimeout(5 * time.Minute)
		defer tabCtx.Close()

		// Navigate to the target URL in main tab
		err = tabCtx.Run(
			chromedp.Navigate(url),
			chromedp.Title(&visit.Title),
		)
		if err != nil {
			visit.Success = false
			visit.ErrorMsg = err.Error()
			core.Logger.Errorf("Failed to process URL %s: %v", url, err)
			return
		}

		// Open control page in default browser
		core.Logger.Info("Opening control panel...")

		if err := server.ApplyCurrentUrl(url); err != nil {
			core.Logger.Error("Failed to open control panel:", err)
		}
		server.Wait()

		err = tabCtx.Run(chromedp.OuterHTML("html", &visit.Html))
		if err != nil {
			visit.Success = false
			visit.ErrorMsg = err.Error()
			core.Logger.Errorf("Failed to capture HTML: %v", err)
			return
		}

		// Save to database using the existing postgres package
		if err := postgres.ChromeVisitRepo.Create(ctx, visit); err != nil {
			core.Logger.Errorf("Failed to save to database: %v", err)
		}

		core.Logger.Infof("Processed URL: %s - Title: %s", url, visit.Title)
	})

	core.Logger.Info("All URLs processed")
	core.Logger.Info("Closing control panel...")
}
