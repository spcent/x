package concurrent

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// ConcurrentExecute is a general-purpose concurrent execution function
// Parameters:
// - ctx: Context
// - ids: List of IDs to process
// - task: Function to process a single ID (additional parameters can be captured via closures if needed)
// - maxConcurrent: Maximum concurrency (default 20, adjustable based on business needs)
// Returns: Aggregated error message containing all failed IDs and their errors
func ConcurrentExecute(
	ctx context.Context,
	ids []int64,
	task func(context.Context, int64) error,
	maxConcurrent ...int,
) error {
	// Set default maximum concurrency
	concurrency := 20
	if len(maxConcurrent) > 0 && maxConcurrent[0] > 0 {
		concurrency = maxConcurrent[0]
	}

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	errs := make(map[int64]error)
	var mu sync.Mutex

	for _, id := range ids {
		sem <- struct{}{} // Acquire semaphore
		wg.Add(1)
		go func(taskID int64) {
			defer func() {
				<-sem // Release semaphore
				wg.Done()
			}()

			// Execute a single task
			if err := task(ctx, taskID); err != nil {
				mu.Lock()
				errs[taskID] = err
				mu.Unlock()
			}
		}(id)
	}

	wg.Wait()

	// Aggregate error messages
	if len(errs) > 0 {
		msg := "concurrent execute failed for IDs: "
		for id, err := range errs {
			msg += fmt.Sprintf("%d(%v), ", id, err)
		}
		return errors.New(msg)
	}

	return nil
}
