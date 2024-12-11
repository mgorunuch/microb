package main

import (
	"github.com/mgorunuch/microb/app/core"
	"github.com/mgorunuch/microb/app/engine/binary_edge"
)

func main() {
	core.Init()
	core.ProcessLinesWithCache(core.Config[binary_edge.BinaryEdgeResponse]{
		CacheProvider: core.NewDefaultFileCache[binary_edge.BinaryEdgeResponse]("binary_edge", core.YEAR),
		ThreadsCount:  1,
		KeyFunc:       core.ParseUrlHostName,
		RunFunc:       binary_edge.Run,
	})
}
