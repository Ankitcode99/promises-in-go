package main

import (
	"context"
	"sync"
)

// Promise represents a function that returns a result or an error
type Promise[T any] func(context.Context) (T, error)

// Result represents the outcome of a promise execution
type Result[T any] struct {
	Value T
	Err   error
	Index int
}

func All[T any](ctx context.Context, promises []Promise[T]) ([]T, error) {
	if len(promises) == 0 {
		return []T{}, nil
	}

	// Create channel to receive results
	resultsChan := make(chan Result[T], len(promises))
	var wg sync.WaitGroup

	// Launch each promise in a goroutine
	for i, promise := range promises {
		wg.Add(1)
		go func(index int, p Promise[T]) {
			defer wg.Done()
			value, err := p(ctx)
			resultsChan <- Result[T]{
				Value: value,
				Err:   err,
				Index: index,
			}
		}(i, promise)
	}

	// Close results channel when all promises complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results maintaining order
	results := make([]T, len(promises))
	for result := range resultsChan {
		if result.Err != nil {
			return nil, result.Err // Return first error encountered
		}
		results[result.Index] = result.Value
	}

	return results, nil
}

// AllSettled executes multiple promises concurrently and returns all results
// regardless of whether they succeeded or failed
func AllSettled[T any](ctx context.Context, promises []Promise[T]) []Result[T] {
	if len(promises) == 0 {
		return []Result[T]{}
	}

	resultsChan := make(chan Result[T], len(promises))
	var wg sync.WaitGroup

	for i, promise := range promises {
		wg.Add(1)
		go func(index int, p Promise[T]) {
			defer wg.Done()
			value, err := p(ctx)
			resultsChan <- Result[T]{
				Value: value,
				Err:   err,
				Index: index,
			}
		}(i, promise)
	}

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
