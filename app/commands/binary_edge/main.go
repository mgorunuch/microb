package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/mgorunuch/microb/app/core"
	"github.com/mgorunuch/microb/app/engine/binary_edge"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func processDomainLine(cacheProvider *core.FileCache[binary_edge.BinaryEdgeResponse]) func(string) {
	return func(line string) {
		hostname, err := core.ParseUrlHostName(line)
		if err != nil {
			// Log error instead of panic for better error handling
			println("Error parsing hostname:", err.Error())
			return
		}

		// Get API key from environment variables
		binaryEdgeApiKey := os.Getenv("BINARYEDGE_API_KEY")
		if binaryEdgeApiKey == "" {
			log.Println("BinaryEdge API key is not set in environment variables")
			return
		}

		// Use the binary_edge package to get data
		response, err := binary_edge.Run(binaryEdgeApiKey, hostname)
		if err != nil {
			log.Println("Error running BinaryEdge Search:", err)
			return
		}

		// The caching is now handled by the FileCache implementation
		if err := cacheProvider.AddToCache(hostname, *response); err != nil {
			log.Println("Error adding to cache:", err)
			return
		}

		fmt.Println("Successfully processed domain:", hostname)
	}
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize the file cache provider for BinaryEdge
	cacheProvider := &core.FileCache[binary_edge.BinaryEdgeResponse]{
		Dir:           filepath.Join("cache", "binary_edge"),
		ExpirationTTL: 365 * 24 * time.Hour, // Cache for 1 year
	}

	var wg sync.WaitGroup
	processor := processDomainLine(cacheProvider)
	core.SpawnAllLines(core.RunParallel(&wg, 1, processor))
	wg.Wait()
}
