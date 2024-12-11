package main

import (
	"github.com/mgorunuch/microb/app/core"
	"github.com/mgorunuch/microb/app/engine/crt_sh"
)

func main() {
	core.Init()
	core.ProcessLinesWithCache(core.Config[[]crt_sh.CertData]{
		CacheProvider: core.NewDefaultFileCache[[]crt_sh.CertData]("crt_sh", core.YEAR),
		ThreadsCount:  1,
		KeyFunc:       core.ParseUrlHostName,
		RunFunc:       crt_sh.Get,
	})
}
