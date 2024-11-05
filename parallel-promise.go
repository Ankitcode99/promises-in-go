package main

import (
	"context"
	"runtime"
	"sync"
)

type Job[T any] struct {
	Index   int
	Promise Promise[T]
}

func ParallelAll[T any](ctx context.Context, promises []Promise[T]) ([]T, error) {
	runtime.GOMAXPROCS(runtime.NumCPU()) // Ensure we use all available CPU cores

	if len(promises) == 0 {
		return []T{}, nil
	}

	resultsChan := make(chan Result[T], len(promises))
	var wg sync.WaitGroup

	// Create a worker pool based on CPU cores
	numWorkers := runtime.NumCPU()
	jobs := make(chan Job[T], len(promises))
	// Start worker goroutines
	for w := 0; w < numWorkers; w++ {
		go func() {
			for job := range jobs {
				value, err := job.Promise(ctx)
				resultsChan <- Result[T]{
					Value: value,
					Err:   err,
					Index: job.Index,
				}
				wg.Done()
			}
		}()
	}

	// Send jobs to worker pool
	for i, promise := range promises {
		wg.Add(1)
		jobs <- Job[T]{
			Index:   i,
			Promise: promise,
		}
	}
	close(jobs)

	// Wait for completion
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results maintaining order
	results := make([]T, len(promises))
	for result := range resultsChan {
		if result.Err != nil {
			return nil, result.Err
		}
		results[result.Index] = result.Value
	}

	return results, nil
}

func worker[T any](
	ctx context.Context,
	jobs <-chan Job[T],
	results chan<- Result[T],
	wg *sync.WaitGroup,
) {
	for job := range jobs {
		value, err := job.Promise(ctx)
		results <- Result[T]{
			Value: value,
			Err:   err,
			Index: job.Index,
		}
		wg.Done()
	}
}

func ParallelAllSettled[T any](ctx context.Context, promises []Promise[T]) []Result[T] {
	if len(promises) == 0 {
		return []Result[T]{}
	}

	jobs := make(chan Job[T], len(promises))

	// Set maximum number of CPUs to use
	numWorkers := runtime.NumCPU()
	runtime.GOMAXPROCS(numWorkers)

	// Create buffered channels for jobs and results
	// jobs := make(chan Job[T], len(promises))
	resultsChan := make(chan Result[T], len(promises))
	var wg sync.WaitGroup

	// Start worker pool
	for w := 0; w < numWorkers; w++ {
		go worker(ctx, jobs, resultsChan, &wg)
	}

	// Send jobs to the worker pool
	for i, promise := range promises {
		wg.Add(1)
		jobs <- Job[T]{
			Index:   i,
			Promise: promise,
		}
	}
	close(jobs) // No more jobs will be sent

	// Wait for all jobs to complete and close results channel
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect all results maintaining order
	results := make([]Result[T], len(promises))
	for result := range resultsChan {
		results[result.Index] = result
	}

	return results
}
