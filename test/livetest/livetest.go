// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package livetest

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/pace/bricks/maintenance/log"
)

// TestFunc represents a single test (possibly with sub tests).
type TestFunc func(t *T)

// Test executes the passed tests in the given order (array order).
// The tests are executed in the configured interval until the given
// context is done.
func Test(ctx context.Context, tests []TestFunc) error {
	t := time.NewTicker(cfg.Interval)

	// run test at least once
	testRun(ctx, tests)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			testRun(ctx, tests)
		}
	}
}

func testRun(ctx context.Context, tests []TestFunc) {
	var err error

	// setup logger in context
	testrun := time.Now().Unix()

	for i, test := range tests {
		logger := log.Ctx(log.WithContext(ctx)).With().
			Int64("livetest", testrun).
			Int("test", i+1).Logger()
		ctx = logger.WithContext(ctx)

		if err = executeTest(ctx, test, fmt.Sprintf("test-%d", i+1)); err != nil {
			break
		}
	}

	if err != nil {
		log.Errorf("failed to run tests: %v", err)
	}
}

func executeTest(ctx context.Context, t TestFunc, name string) error {
	// setup tracing
	span := sentry.StartSpan(ctx, "Livetest")
	defer span.Finish()

	ctx = span.Context()

	logger := log.Ctx(ctx)

	proxy := NewTestProxy(ctx, name)
	startTime := time.Now()

	func() {
		defer func() {
			err := recover()
			if err == nil {
				return
			}

			recoveredErr, ok := err.(error)
			if !ok {
				return
			}

			if !errors.Is(recoveredErr, ErrSkipNow) || !errors.Is(recoveredErr, ErrFailNow) {
				logger.Error().Msgf("PANIC: %+v", err)
				log.Stack(ctx)
				proxy.Fail()
			}
		}()

		t(proxy)
	}()

	duration := float64(time.Since(startTime)) / float64(time.Second)

	proxy.okIfNoSkipFail()
	paceLivetestDurationSeconds.WithLabelValues(cfg.ServiceName).Observe(duration)

	switch proxy.state {
	case StateFailed:
		logger.Warn().Msg("Test failed!")
		span.SetData("test", "failed")
		paceLivetestTotal.WithLabelValues(cfg.ServiceName, "failed").Add(1)
	case StateOK:
		logger.Info().Msg("Test succeeded!")
		span.SetData("test", "succeeded")
		paceLivetestTotal.WithLabelValues(cfg.ServiceName, "succeeded").Add(1)
	case StateSkipped:
		logger.Info().Msg("Test skipped!")
		span.SetData("test", "skipped")
		paceLivetestTotal.WithLabelValues(cfg.ServiceName, "skipped").Add(1)
	default:
		panic(fmt.Errorf("invalid state: %v", proxy.state))
	}

	return nil
}
