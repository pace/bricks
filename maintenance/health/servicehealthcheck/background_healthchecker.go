package servicehealthcheck

import (
	"context"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go"

	"github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/log"
)

// registerBackgroundHealthCheck will run the HealthCheck in the background at every given interval.
// The returned backgroundStateHealthChecker is then used to query this cached state when needed.
func registerBackgroundHealthCheck(name string, hc healthCheck) *backgroundStateHealthChecker {
	var bgState ConnectionState

	go func(ctx context.Context) {
		defer errors.HandleWithCtx(ctx, fmt.Sprintf("BackgroundHealthCheck %s", name))

		var initErr error
		if initHC, ok := hc.check.(Initializable); ok {
			initErr = initHealthCheck(ctx, initHC)
			if initErr != nil {
				log.Warnf("error initializing background health check %q: %s", name, initErr)
				bgState.SetErrorState(initErr)
			}
		}
		timer := time.NewTimer(0) // Do first health check instantly
		for {
			<-timer.C
			func() {
				defer errors.HandleWithCtx(ctx, fmt.Sprintf("BackgroundHealthCheck_HealthCheck %s", name))
				defer timer.Reset(hc.runInBackgroundInterval)

				ctx, cancel := context.WithTimeout(ctx, hc.maxWait)
				defer cancel()
				span, ctx := opentracing.StartSpanFromContext(ctx, "BackgroundHealthCheck")
				defer span.Finish()

				// If Init failed, try again
				if initErr != nil {
					if time.Since(bgState.LastChecked()) < hc.initResultErrorTTL {
						// Too soon, leave the same state
						return
					}
					initErr = initHealthCheck(ctx, hc.check.(Initializable))
					if initErr != nil {
						// Init failed again
						bgState.SetErrorState(initErr)
						return
					}

					// Init succeeded, proceed with check
				}

				// Actual health check
				bgState.setConnectionState(hc.check.HealthCheck(ctx))
			}()
		}
	}(context.Background())

	// Return a HealthCheck that just checks the background "cached" state.
	return &backgroundStateHealthChecker{&bgState}
}

// initHealthCheck will recover from panics and return a proper error
func initHealthCheck(ctx context.Context, initHC Initializable) (err error) {
	defer func() {
		if rp := recover(); rp != nil {
			err = fmt.Errorf("panic: %v", rp)
			errors.Handle(ctx, rp)
		}
	}()

	return initHC.Init(ctx)
}

var _ HealthCheck = (*backgroundStateHealthChecker)(nil)

type backgroundStateHealthChecker struct {
	*ConnectionState
}

func (c *backgroundStateHealthChecker) HealthCheck(_ context.Context) HealthCheckResult {
	return c.GetState()
}
