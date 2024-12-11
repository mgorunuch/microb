package core

import "os"

type env struct{}

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
