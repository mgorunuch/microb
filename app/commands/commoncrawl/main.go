package main

import (
	"github.com/mgorunuch/microb/app/core"
	"github.com/mgorunuch/microb/app/core/cache"
	"github.com/mgorunuch/microb/app/engine/commoncrawl"
)

func main() {
	core.Init()
	core.ProcessLinesWithCache(core.Config[[]commoncrawl.CrawlData]{
		CacheProvider: cache.NewDefaultFileCache[[]commoncrawl.CrawlData]("commoncrawl", core.YEAR),
		ThreadsCount:  1,
		KeyFunc:       core.ParseUrlHostName,
		RunFunc:       commoncrawl.Get,
	})
}
