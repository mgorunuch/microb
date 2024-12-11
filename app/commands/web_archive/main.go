package main

import (
	"github.com/mgorunuch/microb/app/core"
	"github.com/mgorunuch/microb/app/engine/web_archive"
)

func main() {
	core.Init()
	core.ProcessLinesWithCache(core.Config[[]string]{
		CacheProvider: core.NewDefaultFileCache[[]string]("web_archive", core.YEAR),
		ThreadsCount:  1,
		KeyFunc:       core.ParseUrlHostName,
		RunFunc:       web_archive.Get,
	})
}
