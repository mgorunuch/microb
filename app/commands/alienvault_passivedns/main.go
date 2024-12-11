package main

import (
	"github.com/mgorunuch/microb/app/core"
	"github.com/mgorunuch/microb/app/engine/alienvault_passivedns"
	"path/filepath"
	"sync"
	"time"
)

func processDomainLine(cacheProvider *core.FileCache[alienvault_passivedns.PassiveDnsResp]) func(string) {
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

		// Use the alienvault_passivedns package to get data
		_, err = alienvault_passivedns.Get(cacheProvider, hostname)
		if err != nil {
			println("Error getting passive DNS data:", err.Error())
			return
		}

		// The caching is now handled by the FileCache implementation
		println("Successfully processed domain:", hostname)
	}
}

func main() {
	core.Init()

	// Initialize the file cache provider
	cacheProvider := &core.FileCache[alienvault_passivedns.PassiveDnsResp]{
		Dir:           filepath.Join("cache", "alienvault_passivedns"),
		ExpirationTTL: 365 * 24 * time.Hour, // Cache for 1 year
	}

	var wg sync.WaitGroup
	processor := processDomainLine(cacheProvider)
	core.SpawnAllLines(core.RunParallel(&wg, 1, processor))
	wg.Wait()
}
