package redis

import (
	"context"
	"time"

	"github.com/pace/bricks/pkg/routine"
)

type Lock struct {
	rl *routine.Lock
}

type Option func() routine.LockOption

func NewLock(name string, opts ...Option) *Lock {
	rOptions := make([]routine.LockOption, 0)
	for _, opt := range opts {
		rOptions = append(rOptions, opt())
	}
	return &Lock{rl: routine.NewLock(name, rOptions...)}
}

func (l *Lock) Acquire(ctx context.Context) (bool, error) {
	return l.rl.Acquire(ctx)
}

func (l *Lock) AcquireWait(ctx context.Context) error {
	return l.rl.AcquireWait(ctx)
}

func (l *Lock) AcquireAndKeepUp(ctx context.Context) (context.Context, context.CancelFunc, error) {
	return l.rl.AcquireAndKeepUp(ctx)
}

func (l *Lock) Release(ctx context.Context) error {
	return l.rl.Release(ctx)
}

func SetTTL(ttl time.Duration) Option {
	return func() routine.LockOption {
		return routine.SetTTL(ttl)
	}
}
