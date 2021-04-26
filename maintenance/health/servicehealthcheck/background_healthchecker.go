// Copyright © 2021 by PACE Telematics GmbH. All rights reserved.
// Created at 2021/04/26 by Florian Schäfer

package servicehealthcheck

import (
	"context"
	"fmt"
	"sync"
	"time"

	brickserrors "github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/log"
)

type backgroundHealthCheckResult struct {
	lastCheckTime time.Time
	lastResult    HealthCheckResult
}

var (
	backgroundCheckResults sync.Map
)

func startBackgroundHealthCheck(rootCtx context.Context, name string, hc HealthCheck, interval time.Duration) {
	go func() {
		// copy the context, so as to avoid memory leaks
		s := log.Sink{Silent: true}
		ctx := log.ContextWithSink(rootCtx, &s)
		defer brickserrors.HandleWithCtx(ctx, fmt.Sprintf("health-check(%s)", name))
		// restart if it fails. Not the most sophisticated solution I guess, but it should work
		defer startBackgroundHealthCheck(rootCtx, name, hc, interval)

		healthTicker := time.NewTicker(interval)
		for range healthTicker.C {
			timedCtx, cancel := context.WithTimeout(ctx, cfg.HealthCheckMaxWait)
			res := hc.HealthCheck(timedCtx)
			entry := &backgroundHealthCheckResult{
				lastCheckTime: time.Now(),
				lastResult:    res,
			}
			backgroundCheckResults.Store(name, entry)
			cancel()
		}
	}()
}

func getBackgroundHealthCheckResult(name string) *backgroundHealthCheckResult {
	r, ok := backgroundCheckResults.Load(name)
	if !ok {
		return nil
	}
	return r.(*backgroundHealthCheckResult)
}

// getRequiredResults returns all the background check results of the required checks
func getRequiredResults() map[string]HealthCheckResult {
	m := make(map[string]HealthCheckResult)
	requiredChecks.Range(func(key interface{}, val interface{}) bool {
		name := key.(string)
		res := getBackgroundHealthCheckResult(name)
		if time.Since(res.lastCheckTime) > cfg.HealthCheckResultTTL {
			m[name] = HealthCheckResult{
				Msg:   fmt.Sprintf("%s: last health check result is too old. Was: %s", name, string(res.lastResult.State)),
				State: Err,
			}
		}
		m[name] = res.lastResult
		return true
	})
	return m
}

// getOptionalREsults returns all the background check results of the optional checks
func getOptionalResults() map[string]HealthCheckResult {
	m := make(map[string]HealthCheckResult)
	optionalChecks.Range(func(key interface{}, val interface{}) bool {
		name := key.(string)
		res := getBackgroundHealthCheckResult(name)
		if time.Since(res.lastCheckTime) > cfg.HealthCheckResultTTL {
			m[name] = HealthCheckResult{
				Msg:   fmt.Sprintf("%s: last health check result is too old. Was: %s", name, string(res.lastResult.State)),
				State: Err,
			}
		}
		m[name] = res.lastResult
		return true
	})
	return m
}
