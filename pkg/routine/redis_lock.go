// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/02/27 by Marius Neugebauer

package routine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bsm/redislock"
	"github.com/go-redis/redis"
	exponential "github.com/jpillora/backoff"
	redisbackend "github.com/pace/bricks/backend/redis"
	"github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/log"
)

func routineThatKeepsRunningOneInstance(name string, routine func(context.Context)) func(context.Context) {
	return func(ctx context.Context) {
		locker := redislock.New(getDefaultRedisClient())

		// The retry interval is used if we did not get the lock because some
		// other caller got it. The exponential backoff is used if we encounter
		// problems with obtaining the lock, like the Redis not being available.
		// The retry interval is also used if the routine returned regularly, to
		// avoid uncontrollably short restart cycles. If the routine panicked we
		// use exponential backoff as well.
		retryInterval := cfg.RedisLockTTL / 5
		backoff := exponential.Backoff{
			Min: retryInterval,
			Max: 10 * time.Minute,
		}
		routineBackoff := backoff

		num := ctx.Value(ctxNumKey{}).(int64)
		var tryAgainIn time.Duration // zero on first run
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(tryAgainIn):
			}
			// lockCtx will be a child of singleRunCtx. Make sure to cancel the
			// singleRunCtx so that the lock is released after the routine
			// returned.
			singleRunCtx, cancel := context.WithCancel(ctx)
			lockCtx, err := obtainLock(singleRunCtx, locker, "routine:lock:"+name, cfg.RedisLockTTL)
			if err != nil {
				go errors.Handle(singleRunCtx, err) // report error to Sentry, non-blocking
				cancel()
				routineBackoff.Reset()
				tryAgainIn = backoff.Duration()
				continue
			} else if lockCtx != nil {
				routinePanicked := true
				func() {
					defer errors.HandleWithCtx(singleRunCtx, fmt.Sprintf("routine %d", num)) // handle panics
					routine(lockCtx)
					routinePanicked = false
				}()
				if routinePanicked {
					cancel()
					backoff.Reset()
					tryAgainIn = routineBackoff.Duration()
					continue
				}
			}
			cancel()
			backoff.Reset()
			routineBackoff.Reset()
			tryAgainIn = retryInterval
		}
	}
}

var (
	initRedisOnce sync.Once
	redisClient   *redis.Client
)

func getDefaultRedisClient() *redis.Client {
	initRedisOnce.Do(func() { redisClient = redisbackend.Client() })
	return redisClient
}

// Try to obtain a lock. Return a sub-context of ctx that is canceled once the
// lock is lost or ctx is done.
func obtainLock(ctx context.Context, locker *redislock.Client, key string, ttl time.Duration) (context.Context, error) {
	num := ctx.Value(ctxNumKey{}).(int64)

	// obtain lock
	lock, err := locker.Obtain(key, ttl, nil)
	if err == redislock.ErrNotObtained {
		return nil, nil
	} else if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("could not obtain lock")
		return nil, err
	}

	// keep up lock, cancel lockCtx otherwise
	lockCtx, cancel := context.WithCancel(ctx)
	go func() {
		defer errors.HandleWithCtx(ctx, fmt.Sprintf("routine %d: keep up lock", num)) // handle panics
		defer cancel()
		keepUpLock(ctx, lock, ttl)
		err := lock.Release()
		if err != nil && err != redislock.ErrLockNotHeld {
			log.Ctx(ctx).Debug().Err(err).Msg("could not release lock")
		}
	}()

	return lockCtx, nil
}

// Try to keep up a lock for as long as the context is valid. Return once the
// lock is lost or the context is done.
func keepUpLock(ctx context.Context, lock *redislock.Lock, refreshTTL time.Duration) {
	refreshInterval := refreshTTL / 5
	lockRunsOutIn := refreshTTL // initial value after obtaining the lock
	for {
		select {
		case <-ctx.Done():
			return

		// Return if the lock runs out and was not refreshed. lockRunsOutIn is
		// always greater than refreshInterval, except the last refresh failed.
		case <-time.After(lockRunsOutIn):
			return

		// Try to refresh lock.
		case <-time.After(refreshInterval):
		}
		if err := lock.Refresh(refreshTTL, nil); err == redislock.ErrNotObtained {
			// Don't return just yet. Get the TTL of the lock and try to
			// refresh for as long as the TTL is not over.
			if lockRunsOutIn, err = lock.TTL(); err != nil {
				log.Ctx(ctx).Debug().Err(err).Msg("could not get ttl of lock")
				return // assuming we lost the lock
			}
			continue
		} else if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("could not refresh lock")
			return // assuming we lost the lock
		}
		// reset, because the lock was refreshed
		lockRunsOutIn = refreshTTL
	}
}
