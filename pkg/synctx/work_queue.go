// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/10/12 by Vincent Landgraf

package synctx

import (
	"context"
	"fmt"
	"sync"
)

type WorkFunc func(ctx context.Context) error

type WorkQueue struct {
	wg     WaitGroup
	mu     sync.Mutex
	ctx    context.Context
	done   chan struct{}
	err    error
	cancel func()
}

func NewWorkQueue(ctx context.Context) *WorkQueue {
	ctx, cancel := context.WithCancel(ctx)
	return &WorkQueue{
		ctx:    ctx,
		done:   make(chan struct{}),
		cancel: cancel,
	}
}

func (queue *WorkQueue) Add(description string, fn WorkFunc) {
	queue.wg.Add(1)
	go func() {
		err := fn(queue.ctx)
		// if one of the work queue items fails the whole
		// queue will be canceled
		if err != nil {
			queue.setErr(fmt.Errorf("Failed to %s: %v", description, err))
			queue.cancel()
		}
		queue.wg.Done()
	}()
}

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

func (queue *WorkQueue) Err() error {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	return queue.err
}

// setErr sets the error on the queue if not set already
func (queue *WorkQueue) setErr(err error) {
	if queue.err == nil {
		queue.mu.Lock()
		if queue.err == nil {
			queue.err = err
		}
		queue.mu.Unlock()
	}
}
