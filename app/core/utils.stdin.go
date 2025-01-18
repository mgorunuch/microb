package core

import (
	"bufio"
	"context"
	"os"
	"sync"
	"time"
)

func ReadAllLines(processor func(string)) {
	reader := bufio.NewReader(os.Stdin)

	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				break
			}

			panic(err)
		}

		processor(text)
	}
}

func SpawnAllLines(ch chan string) {
	ReadAllLines(func(line string) {
		ch <- line
	})
	close(ch)
}

func SpawnArrayElements[T any](ch chan T, arr []T) {
	for _, el := range arr {
		ch <- el
	}
	close(ch)
}

type Config[T any] struct {
	CacheProvider CacheProvider[T]
	ThreadsCount  int
	KeyFunc       func(ctx context.Context, val string) (string, error)
	RunFunc       func(ctx context.Context, val string) (T, error)
	OutputFunc    func(T)
	SleepTime     time.Duration
	Ctx           context.Context
}

func ProcessLinesWithCache[T any](config Config[T]) {
	var wg sync.WaitGroup
	SpawnAllLines(RunParallel(&wg, config.ThreadsCount, func(v string) {
		defer func() {
			time.Sleep(config.SleepTime)
		}()

		key, err := config.KeyFunc(config.Ctx, v)
		if err != nil {
			Logger.Errorf("Error parsing key: %s", err.Error())
			return
		}

		isCached, err := config.CacheProvider.HasCached(key)
		if err != nil {
			Logger.Errorf("Error checking cache: %s", err.Error())
			return
		}

		if isCached {
			Logger.Debugf("Already processed: %s", key)
			return
		}

		response, err := config.RunFunc(config.Ctx, key)
		if err != nil {
			Logger.Errorf("Error running: %s", err.Error())
			return
		}

		err = config.CacheProvider.AddToCache(key, response)
		if err != nil {
			Logger.Errorf("Error adding to cache: %s", err.Error())
			return
		}

		Logger.Debugf("Successfully processed: %s", key)
	}))
	wg.Wait()
}

type SimpleConfig[T any] struct {
	Ctx          context.Context
	ThreadsCount int
	KeyFunc      func(ctx context.Context, val string) (string, error)
	RunFunc      func(ctx context.Context, val string) (T, error)
	OutputFunc   func(T)
	SleepTime    time.Duration
	Unique       bool
}

func ProcessLines[T any](config SimpleConfig[T]) {
	var wg sync.WaitGroup
	var uniqueKeys sync.Map

	if config.ThreadsCount == 0 {
		config.ThreadsCount = 10
	}

	if config.KeyFunc == nil {
		config.KeyFunc = func(_ context.Context, s string) (string, error) {
			return s, nil
		}
	}

	if config.SleepTime == 0 {
		config.SleepTime = 10 * time.Millisecond
	}

	SpawnAllLines(RunParallel(&wg, config.ThreadsCount, func(v string) {
		defer func() {
			time.Sleep(config.SleepTime)
		}()

		key, err := config.KeyFunc(config.Ctx, v)
		if err != nil {
			Logger.Errorf("Error parsing key: %s", err.Error())
			return
		}

		if config.Unique {
			if _, exists := uniqueKeys.LoadOrStore(key, struct{}{}); exists {
				Logger.Debugf("Skipping duplicate key: %s", key)
				return
			}
		}

		response, err := config.RunFunc(config.Ctx, key)
		if err != nil {
			Logger.Errorf("Error running: %s", err.Error())
			return
		}

		if config.OutputFunc != nil {
			config.OutputFunc(response)
		}

		Logger.Debugf("Successfully processed: %s", key)
	}))
	wg.Wait()
}
