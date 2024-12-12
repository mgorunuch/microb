package main

import (
	"github.com/mgorunuch/microb/app/core"
	"github.com/mgorunuch/microb/app/core/cache"
	"github.com/mgorunuch/microb/app/engine/alienvault_passivedns"
)

func main() {
	core.Init()
	core.ProcessLinesWithCache(core.Config[alienvault_passivedns.PassiveDnsResp]{
		CacheProvider: cache.NewFileCache[alienvault_passivedns.PassiveDnsResp](core.CommandAlienvaultPassivednsCacheDir, core.YEAR),
		ThreadsCount:  1,
		KeyFunc:       core.ParseUrlHostName,
		RunFunc:       alienvault_passivedns.Get,
	})
}
