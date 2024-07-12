// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package synctx

import (
	"context"
	"fmt"
	"sync"
)

// WorkFunc a function that receives an context and optionally returns
// an error. Returning an error will cancel all other worker functions
type WorkFunc func(ctx context.Context) error

// WorkQueue is a work queue implementation that respects cancellation
// using contexts
type WorkQueue struct {
	wg     WaitGroup
	mu     sync.Mutex
	ctx    context.Context
	done   chan struct{}
	err    error
	cancel func()
}

// NewWorkQueue creates a new WorkQueue that respects
// the passed context for cancellation
func NewWorkQueue(ctx context.Context) *WorkQueue {
	ctx, cancel := context.WithCancel(ctx)
	return &WorkQueue{
		ctx:    ctx,
		done:   make(chan struct{}),
		cancel: cancel,
	}
}

// Add add work to the work queue. The passed description
// will be used for the error message, if any. The function
// will be immediately executed.
func (queue *WorkQueue) Add(description string, fn WorkFunc) {
	queue.wg.Add(1)
	go func() {
		err := fn(queue.ctx)
		// if one of the work queue items fails the whole
		// queue will be canceled
		if err != nil {
			queue.setErr(fmt.Errorf("failed to %s: %v", description, err))
			queue.cancel()
		}
		queue.wg.Done()
	}()
}

// Wait waits until all worker functions are done,
// one worker is failing or the context is canceled
func (queue *WorkQueue) Wait() {
	defer queue.cancel()

	select {
	case <-queue.wg.Finish():
	case <-queue.ctx.Done():
		err := queue.ctx.Err()
		// if the queue was canceled and no error was set already
		// store the error
		if err != nil {
			queue.setErr(err)
		}
	}
}

// Err returns the error if one of the work queue items failed
func (queue *WorkQueue) Err() error {
	queue.mu.Lock()
	defer queue.mu.Unlock()

	return queue.err
}

// setErr sets the error on the queue if not set already
func (queue *WorkQueue) setErr(err error) {
	queue.mu.Lock()
	defer queue.mu.Unlock()

	if queue.err == nil {
		queue.err = err
	}
}
