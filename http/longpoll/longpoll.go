package longpoll

import (
	"context"
	"time"
)

// LongPollFunc should return true if long polling did
// resolve, false otherwise. If error is returned, the
// long polling will be canceled. The passed context
// will be guarded by the time budget (deadline) of the
// longpolling request.
type LongPollFunc func(context.Context) (bool, error)

// Config for long polling.
type Config struct {
	// RetryTime time to wait between two retries
	RetryTime time.Duration
	// MinWaitTime min time to wait
	MinWaitTime time.Duration
	// MaxWaitTime max time to wait
	MaxWaitTime time.Duration
}

// Default configuration for http long polling
// wait half a second between retries, min 1 sec and max 60 sec.
var Default = Config{
	RetryTime:   time.Millisecond * 500,
	MinWaitTime: time.Second,
	MaxWaitTime: time.Second * 60,
}

// Until executes the given function fn until duration d is passed or context is canceled.
// The constraints of the Default configuration apply.
func Until(ctx context.Context, d time.Duration, fn LongPollFunc) (ok bool, err error) {
	return Default.LongPollUntil(ctx, d, fn)
}

// LongPollUntil executes the given function fn until duration d is passed or context is canceled.
// If duration is below or above the MinWaitTime, MaxWaitTime from the Config, the values will
// be set to the allowed min/max respectively. Other checking is up to the caller. The resulting time
// budget is communicated via the provided context. This is a defence measure to not have accidental
// long running routines. If no duration is given (0) the long poll will have exactly one execution.
func (c Config) LongPollUntil(ctx context.Context, d time.Duration, fn LongPollFunc) (ok bool, err error) {
	var cancel context.CancelFunc

	if d != 0 {
		if d < c.MinWaitTime {
			d = c.MinWaitTime
		} else if d > c.MaxWaitTime {
			d = c.MaxWaitTime
		}

		ctx, cancel = context.WithTimeout(ctx, d)
		defer cancel()
	}

	for {
		ok, err = fn(ctx)
		if ok || err != nil || d == 0 {
			return
		}

		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(c.RetryTime):
		}
	}
}
