package core

type CacheProvider[T any] interface {
	HasCached(key string) (bool, error)
	GetFromCache(key string) (T, error)
	AddToCache(key string, value T) error
}
