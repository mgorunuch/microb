package main

import (
	"fmt"
	"github.com/mgorunuch/microb/app/core"
	"github.com/mgorunuch/microb/app/engine/google_custom_search"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func processQueryLine(cacheProvider *core.FileCache[google_custom_search.GoogleCustomSearchResponse]) func(string) {
	return func(line string) {
		query := line

		// Get API key and search engine ID from environment variables
		apiKey := os.Getenv("GOOGLE_CUSTOM_SEARCH_API")
		searchEngineID := os.Getenv("GOOGLE_CUSTOM_SEARCH_ENGINE_ID")
		if apiKey == "" || searchEngineID == "" {
			log.Println("API key or search engine ID is not set in environment variables")
			return
		}

		// Check if the query is already cached
		isCached, err := cacheProvider.HasCached(query)
		if err != nil {
			log.Println("Error checking cache:", err)
			return
		}

		if isCached {
			fmt.Println("Query already processed:", query)
			return
		}

		// Use the google_custom_search package to get data
		response, err := google_custom_search.Run(apiKey, searchEngineID, query)
		if err != nil {
			log.Println("Error running Google Custom Search:", err)
			return
		}

		// The caching is now handled by the FileCache implementation
		if err := cacheProvider.AddToCache(query, *response); err != nil {
			log.Println("Error adding to cache:", err)
			return
		}

		fmt.Println("Successfully processed query:", query)
	}
}

func main() {
	core.Init()

	// Initialize the file cache provider
	cacheProvider := &core.FileCache[google_custom_search.GoogleCustomSearchResponse]{
		Dir:           filepath.Join("cache", "google_custom_search"),
		ExpirationTTL: 365 * 24 * time.Hour, // Cache for 1 year
	}

	var wg sync.WaitGroup
	processor := processQueryLine(cacheProvider)
	core.SpawnAllLines(core.RunParallel(&wg, 1, processor))
	wg.Wait()
}
