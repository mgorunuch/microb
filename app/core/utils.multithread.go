package core

import "sync"

func RunParallel[V any](wg *sync.WaitGroup, threads int, processor func(V)) chan V {
	cn := make(chan V)

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for v := range cn {
				processor(v)
			}
		}()
	}

	return cn
}
