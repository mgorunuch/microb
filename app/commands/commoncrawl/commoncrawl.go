package main

import (
	"fmt"
	"github.com/mgorunuch/microb/app/core"
	"github.com/mgorunuch/microb/app/engine/commoncrawl"
	"log"
	"path/filepath"
	"sync"
	"time"
)

func processDomainLine(cacheProvider *core.FileCache[[]commoncrawl.CrawlData]) func(string) {
	return func(line string) {
		hostname, err := core.ParseUrlHostName(line)
		if err != nil {
			log.Println("Error parsing hostname:", err)
			return
		}

		isCached, err := cacheProvider.HasCached(hostname)
		if err != nil {
			log.Println("Error checking cache:", err)
			return
		}

		if isCached {
			fmt.Println("Domain already processed:", hostname)
			return
		}

		// Use the commoncrawl package to get data
		crawlData, err := commoncrawl.Get(hostname)
		if err != nil {
			log.Println("Error fetching Common Crawl data:", err)
			return
		}

		// The caching is now handled by the FileCache implementation
		if err := cacheProvider.AddToCache(hostname, crawlData); err != nil {
			log.Println("Error adding to cache:", err)
			return
		}

		fmt.Println("Successfully processed domain:", hostname)
	}
}

func main() {
	core.Init()

	// Initialize the file cache provider for Common Crawl
	cacheProvider := &core.FileCache[[]commoncrawl.CrawlData]{
		Dir:           filepath.Join("cache", "commoncrawl"),
		ExpirationTTL: 60 * 24 * time.Hour, // Cache for 60 days
	}

	var wg sync.WaitGroup
	processor := processDomainLine(cacheProvider)
	core.SpawnAllLines(core.RunParallel(&wg, 1, processor))
	wg.Wait()
}
