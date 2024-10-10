// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package synctx

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWorkQueueNoTask(t *testing.T) {
	ctx := context.Background()
	q := NewWorkQueue(ctx)
	q.Wait()

	if q.Err() != nil {
		t.Error("expected no error")
	}
}

func TestWorkQueueOneTask(t *testing.T) {
	ctx1 := context.Background()
	q := NewWorkQueue(ctx1)
	q.Add("some work", func(ctx context.Context) error {
		if ctx1 == ctx {
			t.Error("should not directly pass the context")
		}

		return nil
	})

	q.Wait()

	if q.Err() != nil {
		t.Error("expected no error")
	}
}

func TestWorkQueueOneTaskWithErr(t *testing.T) {
	ctx := context.Background()
	q := NewWorkQueue(ctx)
	q.Add("some work", func(ctx context.Context) error {
		return errors.New("Some error")
	})

	q.Wait()

	if q.Err() == nil {
		t.Error("expected error")
		return
	}

	expected := "failed to some work: Some error"
	if q.Err().Error() != expected {
		t.Errorf("expected error %q, got: %q", q.Err().Error(), expected)
	}
}

func TestWorkQueueOneTaskWithCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	q := NewWorkQueue(ctx)
	q.Add("some work", func(ctx context.Context) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	q.Wait()

	if q.Err() == nil {
		t.Error("expected error")
		return
	}

	expected := "context canceled"
	if q.Err().Error() != expected {
		t.Errorf("expected error %q, got: %q", q.Err().Error(), expected)
	}
}
