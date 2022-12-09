// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2021/04/12 by Florian Schäfer
// Locking code is taken in part from github.com/pace/bricks/pkg/routine/redis_lock.go@v0.1.69

package redis

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v7"
	redisbackend "github.com/pace/bricks/backend/redis"
	pberrors "github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/pkg/routine"
	"github.com/rs/zerolog/log"
)

var (
	initOnce      sync.Once
	redisClient   *redis.Client
	defaultLocker *redislock.Client
)

var (
	ErrCouldNotLock    = errors.New("lock could not be obtained")
	ErrCouldNotRelease = errors.New("lock could not be released")
)

type Lock struct {
	Name string

	locker  *redislock.Client
	lockTTL time.Duration

	lock  *redislock.Lock
	mutex sync.Mutex
}

type LockOption func(l *Lock)

func NewLock(name string, opts ...LockOption) *Lock {
	initClient()
	l := &Lock{Name: name}
	for _, opt := range []LockOption{ // default options
		SetTTL(5 * time.Second),
		SetClient(getDefaultLocker()),
	} {
		opt(l)
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

func (l *Lock) Acquire(ctx context.Context) (bool, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	opts := &redislock.Options{
		RetryStrategy: redislock.NoRetry(),
		Context:       ctx,
	}

	lock, err := l.locker.Obtain(l.Name, l.lockTTL, opts)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Str("lockName", l.Name).Msg("Could not acquire lock")
		switch {
		case errors.Is(err, redislock.ErrNotObtained):
			return false, nil
		default:
			return false, pberrors.Hide(ctx, err, ErrCouldNotLock)
		}
	}

	l.lock = lock
	return true, nil
}

func (l *Lock) AcquireWait(ctx context.Context) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	opts := &redislock.Options{
		RetryStrategy: redislock.LinearBackoff(1 * time.Second),
		Context:       ctx,
	}

	lock, err := l.locker.Obtain(l.Name, l.lockTTL, opts)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Str("lockName", l.Name).Msg("Could not acquire lock")
		return pberrors.Hide(ctx, err, ErrCouldNotLock)
	}

	l.lock = lock
	return nil
}

// AcquireAndKeepUp will acquire a lock, and keep it up constantly until cancel is called,
// the returned context is a lock context and is detached from the parent context, meaning that
// any cancellation/timeout on the parent context does not affect this lock context.
func (l *Lock) AcquireAndKeepUp(ctx context.Context) (context.Context, context.CancelFunc, error) {
	opts := &redislock.Options{
		RetryStrategy: redislock.NoRetry(),
		Context:       ctx,
	}

	lock, err := l.locker.Obtain(l.Name, l.lockTTL, opts)
	if err != nil {
		switch {
		case errors.Is(err, redislock.ErrNotObtained):
			return nil, nil, nil
		default:
			return nil, nil, pberrors.Hide(ctx, err, ErrCouldNotLock)
		}
	}

	// Keep up lock, cancel lockCtx otherwise.
	// To pass the detached context from the goroutine we need to pass it through a channel in the parent goroutine
	lockCtxChan := make(chan context.Context)
	cancel := routine.Run(ctx, func(ctx context.Context) {
		defer pberrors.HandleWithCtx(ctx, fmt.Sprintf("keep up lock %q", l.Name)) // handle panics

		lockCtxChan <- ctx

		keepUpLock(ctx, lock, l.lockTTL)
		err := lock.Release()
		if err != nil && err != redislock.ErrLockNotHeld {
			log.Ctx(ctx).Debug().Err(err).Msgf("could not release lock %q", l.Name)
		}
	})

	return <-lockCtxChan, cancel, nil
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

func (l *Lock) Release(ctx context.Context) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.lock == nil {
		log.Ctx(ctx).Debug().Msg("tried to unlock a lock that does not exist")
		return nil
	}

	if err := l.lock.Release(); err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("error releasing redis lock")
		switch {
		case errors.Is(err, redislock.ErrLockNotHeld):
			// well, since our only goal is that the lock is released, this will suffice
		default:
			return pberrors.Hide(ctx, err, ErrCouldNotRelease)
		}
	}

	l.lock = nil
	return nil
}

func initClient() {
	initOnce.Do(func() {
		redisClient = redisbackend.Client()
		defaultLocker = redislock.New(redisClient)
	})
}

func getDefaultLocker() *redislock.Client {
	initClient()
	return defaultLocker
}

func SetTTL(ttl time.Duration) LockOption {
	return func(l *Lock) {
		l.lockTTL = ttl
	}
}

func SetClient(client *redislock.Client) LockOption {
	return func(l *Lock) {
		l.locker = client
	}
}
