package pipeline

import (
	"context"
	"sync"

	"morphr/internal/config"
)

type ConvertFunc func(ctx context.Context, entry FileEntry, cfg *config.Config) error

type Result struct {
	Entry FileEntry
	Err   error
}

// RunPool processes files concurrently using a fixed-size worker pool.
// The progressFn callback is called (from a single goroutine) after each file completes.
func RunPool(ctx context.Context, files []FileEntry, cfg *config.Config, workers int, fn ConvertFunc, progressFn func(Result)) error {
	if workers <= 0 {
		workers = 1
	}

	jobs := make(chan FileEntry, workers*2)
	results := make(chan Result, workers*2)

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for entry := range jobs {
				select {
				case <-ctx.Done():
					results <- Result{Entry: entry, Err: ctx.Err()}
					return
				default:
					err := fn(ctx, entry, cfg)
					results <- Result{Entry: entry, Err: err}
				}
			}
		}()
	}

	go func() {
		for _, f := range files {
			select {
			case <-ctx.Done():
				break
			case jobs <- f:
			}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var firstErr error
	for r := range results {
		if progressFn != nil {
			progressFn(r)
		}
		if r.Err != nil && firstErr == nil {
			firstErr = r.Err
		}
	}

	return firstErr
}
