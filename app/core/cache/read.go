package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

type ReadAllFileCacheFilesOpts[T any] struct {
	Dir     string
	Process func(rec FileCacheRecord[T]) error
}

type FileCacheRecord[T any] struct {
	Key      string
	Ts       time.Time
	FilePath string
}

func (fcr FileCacheRecord[T]) Read() (res T, err error) {
	data, err := os.ReadFile(fcr.FilePath)
	if err != nil {
		return res, fmt.Errorf("failed to read file: %w", err)
	}

	var resp T
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return res, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return resp, nil
}

func ReadAllFileCacheFiles[T any](opts ReadAllFileCacheFilesOpts[T]) error {
	dirs, err := os.ReadDir(opts.Dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, keyFile := range dirs {
		if !keyFile.IsDir() {
			continue
		}

		files, err := os.ReadDir(fmt.Sprintf("%s/%s", opts.Dir, keyFile.Name()))
		if err != nil {
			return fmt.Errorf("failed to read directory: %w", err)
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			stamp, err := strconv.ParseInt(file.Name(), 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse timestamp: %w", err)
			}

			fcr := FileCacheRecord[T]{
				Key:      keyFile.Name(),
				Ts:       time.Unix(stamp, 0),
				FilePath: fmt.Sprintf("%s/%s/%s", opts.Dir, keyFile.Name(), file.Name()),
			}

			if err := opts.Process(fcr); err != nil {
				return fmt.Errorf("failed to process file: %w", err)
			}
		}
	}

	return nil
}
