package main

import (
	"github.com/mgorunuch/microb/app/core"
	"github.com/mgorunuch/microb/app/engine/google_custom_search"
)

func main() {
	core.Init()
	core.ProcessLinesWithCache(core.Config[google_custom_search.GoogleCustomSearchResponse]{
		CacheProvider: core.NewDefaultFileCache[google_custom_search.GoogleCustomSearchResponse]("google_custom_search", core.YEAR),
		ThreadsCount:  1,
		KeyFunc: func(s string) (string, error) {
			return s, nil
		},
		RunFunc: google_custom_search.Run,
	})
}
