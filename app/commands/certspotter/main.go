package main

import (
	"fmt"
	"github.com/mgorunuch/microb/app/core"
	"github.com/mgorunuch/microb/app/engine/certspotter"
	"log"
	"path/filepath"
	"sync"
	"time"
)

func processDomainLine(cacheProvider *core.FileCache[[]certspotter.Issuance]) func(string) {
	return func(line string) {
		hostname, err := core.ParseUrlHostName(line)
		if err != nil {
			// Log error instead of panic for better error handling
			println("Error parsing hostname:", err.Error())
			return
		}

		// Check if the domain is already cached
		isCached, err := cacheProvider.HasCached(hostname)
		if err != nil {
			println("Error checking cache:", err.Error())
			return
		}

		if isCached {
			println("Domain already processed:", hostname)
			return
		}

		// Use the certspotter package to get data
		issuances, err := certspotter.Get(hostname)
		if err != nil {
			log.Println("Error fetching CertSpotter issuances:", err)
			return
		}

		// The caching is now handled by the FileCache implementation
		if err := cacheProvider.AddToCache(hostname, issuances); err != nil {
			log.Println("Error adding to cache:", err)
			return
		}

		fmt.Println("Successfully processed domain:", hostname)
	}
}

func main() {
	core.Init()

	// Initialize the file cache provider for CertSpotter
	cacheProvider := &core.FileCache[[]certspotter.Issuance]{
		Dir:           filepath.Join("cache", "certspotter"),
		ExpirationTTL: 365 * 24 * time.Hour, // Cache for 1 year
	}

	var wg sync.WaitGroup
	processor := processDomainLine(cacheProvider)
	core.SpawnAllLines(core.RunParallel(&wg, 1, processor))
	wg.Wait()
}
