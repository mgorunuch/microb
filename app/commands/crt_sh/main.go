package main

import (
	"fmt"
	"github.com/mgorunuch/microb/app/core"
	"github.com/mgorunuch/microb/app/engine/crt_sh"
	"log"
	"path/filepath"
	"sync"
	"time"
)

func processDomainLine(cacheProvider *core.FileCache[[]crt_sh.CertData]) func(string) {
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

		// Use the crt_sh package to get data
		certData, err := crt_sh.Get(hostname)
		if err != nil {
			log.Println("Error fetching certificate data:", err)
			return
		}

		// The caching is now handled by the FileCache implementation
		if err := cacheProvider.AddToCache(hostname, certData); err != nil {
			log.Println("Error adding to cache:", err)
			return
		}

		fmt.Println("Successfully processed domain:", hostname)
	}
}

func main() {
	core.Init()

	// Initialize the file cache provider for crt.sh
	cacheProvider := &core.FileCache[[]crt_sh.CertData]{
		Dir:           filepath.Join("cache", "crt_sh"),
		ExpirationTTL: 365 * 24 * time.Hour, // Cache for 1 year
	}

	var wg sync.WaitGroup
	processor := processDomainLine(cacheProvider)
	core.SpawnAllLines(core.RunParallel(&wg, 1, processor))
	wg.Wait()
}
