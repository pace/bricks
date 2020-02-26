// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/02/27 by Marius Neugebauer

package routine

import (
	"context"
	"sync"
	"time"

	"github.com/bsm/redislock"
	"github.com/go-redis/redis"
	redisbackend "github.com/pace/bricks/backend/redis"
)

func routineThatKeepsRunningOneInstance(name string, routine func(context.Context)) func(context.Context) {
	return func(ctx context.Context) {
		locker := redislock.New(getDefaultRedisClient())
		var tryAgainIn time.Duration // zero on first run
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(tryAgainIn):
			}
			lockCtx := obtainLock(ctx, locker, "routine:lock:"+name, cfg.RedisLockTTL)
			if lockCtx != nil {
				routine(lockCtx)
			}
			tryAgainIn = cfg.RedisLockTTL / 5
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
func obtainLock(ctx context.Context, locker *redislock.Client, key string, ttl time.Duration) context.Context {
	// obtain lock
	lock, err := locker.Obtain(key, ttl, nil)
	if err == redislock.ErrNotObtained {
		return nil
	} else if err != nil {
		panic(err)
	}

	// keep up lock, cancel lockCtx otherwise
	lockCtx, cancel := context.WithCancel(ctx)
	go func() {
		keepUpLock(ctx, lock, ttl)
		cancel()
		err := lock.Release()
		if err != nil && err != redislock.ErrLockNotHeld {
			panic(err)
		}
	}()

	return lockCtx
}

// Try to keep up a lock for as long as the context is valid. Return once the
// lock is lost or the context is done.
func keepUpLock(ctx context.Context, lock *redislock.Lock, refreshTTL time.Duration) {
	refreshInterval := refreshTTL / 5
	lockRunsOutIn := refreshTTL // initial value must be > refresh interval
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
				panic(err)
			}
			continue
		} else if err != nil {
			panic(err)
		}
		// reset, because the lock was refreshed
		lockRunsOutIn = refreshTTL
	}
}
