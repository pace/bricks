// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/18 Charlotte Pröller

package servicehealthcheck

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"

	"github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/log"
)

// HealthCheck is a health check that is registered once and that is performed
// periodically in the background.
type HealthCheck interface {
	HealthCheck(ctx context.Context) HealthCheckResult
}

type HealthCheckFunc func(ctx context.Context) HealthCheckResult

func (hcf HealthCheckFunc) HealthCheck(ctx context.Context) HealthCheckResult {
	return hcf(ctx)
}

// Initializable is used to mark that a health check needs to be initialized
type Initializable interface {
	Init(ctx context.Context) error
}

// HealthState describes if a any error or warning occurred during the health check of a service
type HealthState string

const (
	// Err State of a service, if an error occurred during the health check of the service
	Err HealthState = "ERR"
	// Warn State of a service, if a warning occurred during the health check of the service
	Warn HealthState = "WARN"
	// Ok State of a service, if no warning or error occurred during the health check of the service
	Ok HealthState = "OK"
)

// HealthCheckResult describes the result of a health check, contains the state of a service and a message that
// describes the state. If the state is Ok the description can be empty.
// The description should contain the error message if any error or warning occurred during the health check.
type HealthCheckResult struct {
	State HealthState
	Msg   string
}

// requiredChecks contains all required registered Health Checks - key:Name
var requiredChecks sync.Map

// optionalChecks contains all optional registered Health Checks - key:Name
var optionalChecks sync.Map

func checksResults(checks *sync.Map) map[string]HealthCheckResult {
	results := make(map[string]HealthCheckResult)
	checks.Range(func(key, value interface{}) bool {
		name := key.(string)
		result := value.(*ConnectionState).GetState()
		results[name] = result
		return true
	})
	return results
}

// RegisterHealthCheck registers a required HealthCheck. The name
// must be unique. If the health check satisfies the Initializable interface, it
// is initialized before it is added.
// It is not possible to add a health check with the same name twice, even if one is required and one is optional
func RegisterHealthCheck(name string, hc HealthCheck, opts ...HealthCheckOption) {
	registerHealthCheck(&requiredChecks, name, hc, opts...)
}

// RegisterHealthCheckFunc registers a required HealthCheck. The name
// must be unique.  It is not possible to add a health check with the same name twice,
// even if one is required and one is optional
func RegisterHealthCheckFunc(name string, f HealthCheckFunc, opts ...HealthCheckOption) {
	RegisterHealthCheck(name, f, opts...)
}

// RegisterOptionalHealthCheck registers a HealthCheck like RegisterHealthCheck(hc HealthCheck, name string)
// but the health check is only checked for /health/check and not for /health/
func RegisterOptionalHealthCheck(hc HealthCheck, name string, opts ...HealthCheckOption) {
	registerHealthCheck(&optionalChecks, name, hc, opts...)
}

// registerHealthCheck will run the HealthCheck in the background.
func registerHealthCheck(checks *sync.Map, name string, check HealthCheck, opts ...HealthCheckOption) {
	ctx := log.Logger().WithContext(context.Background())
	ctx = log.ContextWithSink(ctx, log.NewSink(log.Silent()))

	// create config based on defaults, then overwrite with given options
	hcCfg := HealthCheckCfg{
		interval:           cfg.Interval,
		initResultErrorTTL: cfg.HealthCheckInitResultErrorTTL,
		maxWait:            cfg.HealthCheckMaxWait,
		warmupDelay:        cfg.HealthCheckWarmupDelay,
	}
	for _, o := range opts {
		o(&hcCfg)
	}

	// check both lists, because
	if _, inReq := requiredChecks.Load(name); inReq {
		log.Warnf("tried to register health check with name %q twice", name)
		return
	}
	if _, inOpt := optionalChecks.Load(name); inOpt {
		log.Warnf("tried to register health check with name %q twice", name)
		return
	}

	// save the length of the longest health check name, for the width of the column in /health/check
	if len(name) > longestCheckName {
		longestCheckName = len(name)
	}
	var bgState ConnectionState
	checks.Store(name, &bgState)

	go func() {
		defer errors.HandleWithCtx(ctx, fmt.Sprintf("BackgroundHealthCheck %s", name))

		var (
			initHC, hasInitialization = check.(Initializable)
			initialized               = false
			warmupFinished            = false
		)
		// Start first health check run instantly
		timer := time.NewTimer(0)
		// calculate when the warmup phase should be finished
		healthCheckStart := time.Now()
		warmupDeadline := healthCheckStart.Add(hcCfg.warmupDelay)
		for {
			<-timer.C
			func() {
				defer errors.HandleWithCtx(ctx, fmt.Sprintf("BackgroundHealthCheck_HealthCheck %s", name))
				defer timer.Reset(hcCfg.interval)

				ctx, cancel := context.WithTimeout(ctx, hcCfg.maxWait)
				defer cancel()
				span, ctx := opentracing.StartSpanFromContext(ctx, "BackgroundHealthCheck")
				defer span.Finish()

				if hasInitialization && !initialized {
					if time.Since(bgState.LastChecked()) < hcCfg.initResultErrorTTL {
						// Too soon, leave the same state
						return
					}
					initErr := initHealthCheck(ctx, initHC)
					if initErr != nil {
						// Init failed again
						bgState.SetErrorState(initErr)
						return
					}

					initialized = true
					// Init succeeded, proceed with check
				}

				// don't execute the first healtcheck before we finished the warmup period
				if !warmupFinished {
					if warmupDeadline.Before(time.Now()) {
						warmupFinished = true
					} else {
						bgState.setConnectionState(HealthCheckResult{
							State: Ok,
							Msg:   fmt.Sprintf("Service warms up since '%s'", healthCheckStart.Format(time.RFC3339)),
						})
						// sanity trigger a health check, since we can not guarantee what the real implementation does ...
						go check.HealthCheck(ctx)
						return
					}
				}

				// Actual health check
				bgState.setConnectionState(check.HealthCheck(ctx))
			}()
		}
	}()
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
