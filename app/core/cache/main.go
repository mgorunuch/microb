package cache

import (
	"encoding/json"
	"fmt"
	"github.com/mgorunuch/microb/app/core"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func NewDefaultFileCache[T any](service string, ttl time.Duration) *FileCache[T] {
	return NewFileCache[T](filepath.Join("cache", service), ttl)
}

func NewFileCache[T any](dir string, expirationTTL time.Duration) *FileCache[T] {
	return &FileCache[T]{
		Dir:           dir,
		ExpirationTTL: expirationTTL,
	}
}

type FileCache[T any] struct {
	Dir           string
	ExpirationTTL time.Duration
}

func (fc *FileCache[T]) getKeyDir(key string) string {
	return filepath.Join(fc.Dir, key)
}

func (fc *FileCache[T]) getCacheFilePath(key string, timestamp time.Time) string {
	return filepath.Join(fc.getKeyDir(key), fmt.Sprintf("%d", timestamp.UnixNano()))
}

func (fc *FileCache[T]) getLatestCacheFile(key string) (string, error) {
	keyDir := fc.getKeyDir(key)

	// Read all files in the key directory
	files, err := os.ReadDir(keyDir)
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", os.ErrNotExist
	}

	// Get all timestamps
	var timestamps []int64
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ts, err := core.ParseInt64(file.Name())
		if err != nil {
			continue
		}
		timestamps = append(timestamps, ts)
	}

	if len(timestamps) == 0 {
		return "", os.ErrNotExist
	}

	// Sort timestamps in descending order
	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i] > timestamps[j]
	})

	// Check if the latest file is expired
	latestTime := time.Unix(0, timestamps[0])
	if fc.ExpirationTTL != 0 && time.Since(latestTime) > fc.ExpirationTTL {
		return "", os.ErrNotExist
	}

	return filepath.Join(keyDir, fmt.Sprintf("%d", timestamps[0])), nil
}

func (fc *FileCache[T]) HasCached(key string) (bool, error) {
	_, err := fc.getLatestCacheFile(key)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("error checking cache: %w", err)
	}
	return true, nil
}

func (fc *FileCache[T]) GetFromCache(key string) (T, error) {
	var zero T

	filePath, err := fc.getLatestCacheFile(key)
	if err != nil {
		return zero, fmt.Errorf("error getting latest cache file: %w", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return zero, fmt.Errorf("error reading cache file: %w", err)
	}

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return zero, fmt.Errorf("error unmarshaling cache data: %w", err)
	}

	return value, nil
}

func (fc *FileCache[T]) AddToCache(key string, value T) error {
	keyDir := fc.getKeyDir(key)
	if err := os.MkdirAll(keyDir, 0755); err != nil {
		return fmt.Errorf("error creating cache directory: %w", err)
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("error marshaling cache data: %w", err)
	}

	filePath := fc.getCacheFilePath(key, time.Now())
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("error writing cache file: %w", err)
	}

	return nil
}

func (fc *FileCache[T]) CleanExpired() error {
	if fc.ExpirationTTL == 0 {
		return nil
	}

	return filepath.Walk(fc.Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ts, err := core.ParseInt64(info.Name())
		if err != nil {
			return nil
		}

		cacheTime := time.Unix(0, ts)
		if time.Since(cacheTime) > fc.ExpirationTTL {
			_ = os.Remove(path)
		}

		return nil
	})
}
