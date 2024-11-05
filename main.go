package main

import (
	"context"
	"fmt"
	"math"
	"sync/atomic"
	"time"
)

func concurrentImplementation() {
	ctx := context.Background()

	// Define some example promises
	promises := []Promise[int]{
		func(ctx context.Context) (int, error) {
			time.Sleep(100 * time.Millisecond)
			return 1, nil
		},
		func(ctx context.Context) (int, error) {
			time.Sleep(50 * time.Millisecond)
			return 2, nil
		},
		func(ctx context.Context) (int, error) {
			return -1, fmt.Errorf("promise 3 failed")
		},
	}

	// Using All
	results, err := All(ctx, promises)
	if err != nil {
		fmt.Println("Error:", err)
		// return
	} else {
		fmt.Println("All results:", results)
	}
	// fmt.Println("All results:", results) // [1, 2, 3]

	// Using AllSettled
	settledResults := AllSettled(ctx, promises)
	for i, result := range settledResults {
		if result.Err != nil {
			fmt.Printf("Promise %d failed: %v\n", i, result.Err)
		} else {
			fmt.Printf("Promise %d succeeded: %v\n", i, result.Value)
		}
	}
}

func parallelImplementation() {
	ctx := context.Background()

	// CPU-intensive promises
	promises := make([]Promise[int], 4)
	for i := 0; i < 4; i++ {
		// i := i
		promises[i] = func(ctx context.Context) (int, error) {
			// Simulate CPU-intensive work
			result := 0
			for j := 0; j < 10000000000; j++ {
				result += j
			}
			return i + result, nil
		}
	}

	// Time both implementations
	start := time.Now()
	results1, _ := All(ctx, promises)
	fmt.Println("All results1 - ", results1)
	fmt.Printf("Concurrent execution took: %v\n", time.Since(start))

	start = time.Now()
	results2, _ := ParallelAll(ctx, promises)
	fmt.Println("All results2 - ", results2)
	fmt.Printf("Parallel execution took: %v\n", time.Since(start))

	if len(results1) == len(results2) {
		fmt.Println("Results are the same")
	}
}

var MAX_INT = 100000000
var CONCURRENCY = 10
var totalPrimeNumbers int32 = 0
var currentNum int32 = 2

func checkPrime(x int) {
	if x&1 == 0 {
		return
	}
	for i := 3; i <= int(math.Sqrt(float64(x))); i++ {
		if x%i == 0 {
			return
		}
	}
	atomic.AddInt32(&totalPrimeNumbers, 1)
}

func main() {

	concurrentImplementation()

	fmt.Println("\n\n---- RUNNING PARALLEL IMPLEMENTATION ----")

	parallelImplementation()

}
