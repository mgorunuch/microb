package main

import (
	"github.com/mgorunuch/microb/app/core"
	"github.com/mgorunuch/microb/app/engine/certspotter"
)

func main() {
	core.Init()
	core.ProcessLinesWithCache(core.Config[[]certspotter.Issuance]{
		CacheProvider: core.NewDefaultFileCache[[]certspotter.Issuance]("certspotter", core.YEAR),
		ThreadsCount:  1,
		KeyFunc:       core.ParseUrlHostName,
		RunFunc:       certspotter.Get,
	})
}
