package longpoll

import (
	"context"
	"time"
)

// LongPollFunc should return true if long polling did
// resolve, false otherwise. If error is returned, the
// long polling will be canceled
type LongPollFunc func() (bool, error)

// Config for long polling
type Config struct {
	// RetryTime time to wait between two retries
	RetryTime time.Duration
	// MinWaitTime min time to wait
	MinWaitTime time.Duration
	// MaxWaitTime max time to wait
	MaxWaitTime time.Duration
}

// Default configuration for http long polling
// wait half a second between retries, min 1 sec and max 60 sec
var Default = Config{
	RetryTime:   time.Millisecond * 500,
	MinWaitTime: time.Second,
	MaxWaitTime: time.Second * 60,
}

// LongPollUntil executes the given function fn until duration d is passed or context is canceled
func LongPollUntil(ctx context.Context, d time.Duration, fn LongPollFunc) (ok bool, err error) {
	return Default.LongPollUntil(ctx, d, fn)
}

// LongPollUntil executes the given function fn until duration d is passed or context is canceled
func (c Config) LongPollUntil(ctx context.Context, d time.Duration, fn LongPollFunc) (ok bool, err error) {
	until := time.Now()
	if d >= c.MinWaitTime && d < c.MaxWaitTime {
		until = until.Add(d)
	}

loop:
	for {
		ok, err = fn()
		if err != nil {
			return
		}

		// fn returns true, break the loop
		if ok {
			break
		}

		// no long polling
		if d <= 0 {
			break
		}

		// long pooling desired?
		if !time.Now().Before(until) {
			break
		}

		select {
		// handle context cancelation
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(c.RetryTime):
			continue loop
		}
	}

	return
}
