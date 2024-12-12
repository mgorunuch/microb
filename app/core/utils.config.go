package core

import (
	"os"
	"path"
)

func cacheDir(name string) string {
	return path.Join("cache", name)
}

var (
	CommandAlienvaultPassivedns         = `alienvault_passivedns`
	CommandAlienvaultPassivednsCacheDir = cacheDir(CommandAlienvaultPassivedns)

	CommandBinaryEdge         = "binary_edge"
	CommandBinaryEdgeCacheDir = cacheDir(CommandBinaryEdge)

	CommandCertspotter         = "certspotter"
	CommandCertspotterCacheDir = cacheDir(CommandCertspotter)

	CommandCommonCrawl         = "commoncrawl"
	CommandCommonCrawlCacheDir = cacheDir(CommandCommonCrawl)

	CommandCrtSh         = "crt_sh"
	CommandCrtShCacheDir = cacheDir(CommandCrtSh)

	CommandGoogleSearch         = "google_custom_search"
	CommandGoogleSearchCacheDir = cacheDir(CommandGoogleSearch)

	CommandWebArchive         = "web_archive"
	CommandWebArchiveCacheDir = cacheDir(CommandWebArchive)
)

type env struct{}

func (e env) GetDefault(key, defaultValue string) string {
	val := e.Get(key, false)
	if val == "" {
		return defaultValue
	}

	return val
}

func (e env) Get(key string, required bool) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	if required {
		Logger.Fatalf("Environment variable %s is required", key)
	}

	return ""
}

func (e env) BINARYEDGE_API_KEY() string {
	return e.Get("BINARYEDGE_API_KEY", true)
}

var Env = env{}
