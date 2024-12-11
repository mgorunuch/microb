package core

import "time"

type CacheProvider[T any] interface {
	HasCached(key string) (bool, error)
	GetFromCache(key string) (T, error)
	AddToCache(key string, value T) error
}

var YEAR = time.Hour * 24 * 365
var MONTH = time.Hour * 24 * 30
var WEEK = time.Hour * 24 * 7
