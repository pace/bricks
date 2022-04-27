package servicehealthcheck

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"

	"github.com/pace/bricks/maintenance/log"
)

// BackgroundHealthCheck is an interface that, when implemented, will run the
// HealthCheck in the background, and cache the result, that will be fetched by the HealthHandler
type BackgroundHealthCheck interface {
	HealthCheck
	Interval() time.Duration
}

// BackgroundHealthCheckMaxWait allows to use a custom amount of time to wait before failing the health check
type BackgroundHealthCheckMaxWait interface {
	MaxWait() time.Duration
}

// registerBackgroundHealthCheck will run the HealthCheck in the background at every given interval.
// The returned backgroundStateHealthChecker is then used to query this cached state when needed.
func registerBackgroundHealthCheck(name string, bhc BackgroundHealthCheck) *backgroundStateHealthChecker {
	var bgState ConnectionState

	go func() {
		s := log.Sink{Silent: true}
		ctx := log.ContextWithSink(context.Background(), &s)

		var initErr error
		if initHC, ok := bhc.(Initializable); ok {
			if initErr = initHC.Init(ctx); initErr != nil {
				log.Warnf("error initializing background health check %q: %s", name, initErr)
				bgState.SetErrorState(initErr)
			}
		}
		timer := time.NewTimer(0) // Do first health check instantly
		for {
			<-timer.C
			func() {
				defer timer.Reset(bhc.Interval())

				maxWait := cfg.HealthCheckMaxWait
				if mw, ok := bhc.(BackgroundHealthCheckMaxWait); ok {
					maxWait = mw.MaxWait()
				}
				ctx, cancel := context.WithTimeout(ctx, maxWait)
				defer cancel()
				span, ctx := opentracing.StartSpanFromContext(ctx, "BackgroundHealthCheck")
				defer span.Finish()

				// If Init failed, try again
				if initErr != nil {
					if time.Since(bgState.LastChecked()) < cfg.HealthCheckInitResultErrorTTL {
						// Too soon, leave the same state
						return
					}
					initErr = bhc.(Initializable).Init(ctx)
					if initErr != nil {
						// Init failed again
						bgState.SetErrorState(initErr)
						return
					}

					// Init succeeded, proceed with check
				}

				// Actual health check
				bgState.setConnectionState(bhc.HealthCheck(ctx))
			}()
		}
	}()

	// Return a HealthCheck that just checks the background "cached" state.
	return &backgroundStateHealthChecker{&bgState}
}

var _ HealthCheck = (*backgroundStateHealthChecker)(nil)

type backgroundStateHealthChecker struct {
	*ConnectionState
}

func (c *backgroundStateHealthChecker) HealthCheck(_ context.Context) HealthCheckResult {
	return c.GetState()
}
