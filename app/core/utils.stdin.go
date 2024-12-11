package core

import (
	"bufio"
	"os"
	"sync"
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
	KeyFunc       func(string) (string, error)
	RunFunc       func(string) (T, error)
}

func ProcessLinesWithCache[T any](config Config[T]) {
	var wg sync.WaitGroup
	SpawnAllLines(RunParallel(&wg, config.ThreadsCount, func(v string) {
		key, err := config.KeyFunc(v)
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

		response, err := config.RunFunc(v)
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
