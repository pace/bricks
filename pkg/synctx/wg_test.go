// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package synctx

import (
	"context"
	"testing"
	"time"
)

func TestWaitGroupOk(t *testing.T) {

	var wg1 WaitGroup

	wg1.Add(1)
	wg1.Add(1)

	ctx := context.Background()

	go func() { time.Sleep(time.Millisecond * 5); wg1.Done() }()
	go func() { time.Sleep(time.Millisecond * 5); wg1.Done() }()

	select {
	case <-wg1.Finish():
	case <-ctx.Done():
		if ctx.Err() != nil {
			t.Error("Context should never be done:", ctx.Err())
		}
	}
}

func TestWaitGroupFail(t *testing.T) {

	var wg1 WaitGroup

	wg1.Add(1)
	wg1.Add(1)

	ctx, cancel := context.WithCancel(context.Background())

	go func() { time.Sleep(time.Millisecond * 5); wg1.Done() }()
	go func() { time.Sleep(time.Millisecond * 5); cancel() }()

	select {
	case <-wg1.Finish():
	case <-ctx.Done():
		if ctx.Err() != context.Canceled {
			t.Error("Context should never be done", ctx.Err())
		}
	}
}
