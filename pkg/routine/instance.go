// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package routine

import (
	"context"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/pkg/lock/redis"

	exponential "github.com/jpillora/backoff"
)

type routineThatKeepsRunningOneInstance struct {
	Name    string
	Routine func(context.Context)

	lockTTL       time.Duration
	retryInterval time.Duration
	backoff       combinedExponentialBackoff
	num           int64
}

func (r *routineThatKeepsRunningOneInstance) Run(ctx context.Context) {
	// The retry interval is used if we did not get the lock because some
	// other caller got it. The exponential backoff is used if we encounter
	// problems with obtaining the lock, like the Redis not being available.
	// The retry interval is also used if the routine returned regularly, to
	// avoid uncontrollably short restart cycles. If the routine panicked we
	// use exponential backoff as well.
	r.lockTTL = cfg.RedisLockTTL
	r.retryInterval = r.lockTTL / 5
	r.backoff = combinedExponentialBackoff{
		"lock":    &exponential.Backoff{Min: r.retryInterval, Max: 10 * time.Minute},
		"routine": &exponential.Backoff{Min: r.retryInterval, Max: 10 * time.Minute},
	}

	r.num = ctx.Value(ctxNumKey{}).(int64)
	var tryAgainIn time.Duration // zero on first run
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(tryAgainIn):
		}
		// Make sure to cancel the singleRunCtx so that the lock is released
		// after the routine returned.
		singleRunCtx, cancel := context.WithCancel(ctx)
		tryAgainIn = r.singleRun(singleRunCtx)
		cancel()
	}
}

// Performs a single run. That is, to try to obtain the lock and run the routine
// until it returns. Return the backoff duration after which another single run
// should be performed.
func (r *routineThatKeepsRunningOneInstance) singleRun(ctx context.Context) time.Duration {
	l := redis.NewLock("routine:lock:"+r.Name, redis.SetTTL(r.lockTTL))
	lockCtx, cancel, err := l.AcquireAndKeepUp(ctx)
	if err != nil {
		go errors.Handle(ctx, err) // report error to Sentry, non-blocking
		return r.backoff.Duration("lock")
	}
	if lockCtx != nil {
		defer cancel()
		routinePanicked := true
		func() {
			defer errors.HandleWithCtx(ctx, fmt.Sprintf("routine %d", r.num)) // handle panics

			span := sentry.StartSpan(lockCtx, "function", sentry.WithDescription(fmt.Sprintf("routine %d", r.num)))
			defer span.Finish()

			r.Routine(span.Context())
			routinePanicked = false
		}()
		if routinePanicked {
			return r.backoff.Duration("routine")
		}
	}
	r.backoff.ResetAll()
	return r.retryInterval
}
