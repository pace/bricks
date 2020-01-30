// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/18 Charlotte Pröller

package servicehealthcheck

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/caarlos0/env"
	"github.com/pace/bricks/maintenance/log"
)

// HealthCheck is a health check that is registered once and that is performed
// periodically and/or spontaneously.
type HealthCheck interface {
	HealthCheck() HealthCheckResult
}

// Initialisable is used to mark that a health check needs to be initialised
type Initialisable interface {
	Init() error
}

type config struct {
	// Amount of time to cache the last init
	HealthCheckInitResultErrorTTL time.Duration `env:"HEALTH_CHECK_INIT_RESULT_ERROR_TTL" envDefault:"10s"`
}

var cfg config

// requiredChecks contains all required registered Health Checks - key:Name
var requiredChecks sync.Map

// optionalChecks contains all optional registered Health Checks - key:Name
var optionalChecks sync.Map

// initErrors map with all err ConnectionState that happened in the initialisation of any health check - key:Name
var initErrors sync.Map

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

func init() {
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Failed to parse health check environment: %v", err)
	}
}

func check(hcs *sync.Map) map[string]HealthCheckResult {
	result := make(map[string]HealthCheckResult)
	hcs.Range(func(key, value interface{}) bool {
		name := key.(string)
		hc := value.(HealthCheck)
		// If it was not possible to initialise this health check, then show the initialisation error message
		if val, isIn := initErrors.Load(name); isIn {
			if done := reInitHealthCheck(val.(*ConnectionState), result, name, hc.(Initialisable)); done {
				return true
			}
		}
		// this is the actual health check
		result[name] = hc.HealthCheck()
		return true
	})
	return result
}

func reInitHealthCheck(conState *ConnectionState, result map[string]HealthCheckResult, name string, initHc Initialisable) bool {
	if time.Since(conState.LastChecked()) < cfg.HealthCheckInitResultErrorTTL {
		result[name] = conState.GetState()
		return true
	}
	err := initHc.Init()
	if err != nil {
		conState.SetErrorState(err)
		result[name] = conState.GetState()
		return true
	}
	initErrors.Delete(name)
	return false
}

func writeResult(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	if _, err := fmt.Fprint(w, body); err != nil {
		log.Warnf("could not write output: %s", err)
	}
}

// RegisterHealthCheck registers a required HealthCheck. The name
// must be unique. If the health check satisfies the Initialisable interface, it
// is initialised before it is added.
// It is not possible to add a health check with the same name twice, even if one is required and one is optional
func RegisterHealthCheck(hc HealthCheck, name string) {
	registerHealthCheck(&requiredChecks, hc, name)
}

// RegisterOptionalHealthCheck registers a HealthCheck like RegisterHealthCheck(hc HealthCheck, name string)
// but the health check is only checked for /health/check and not for /health/
func RegisterOptionalHealthCheck(hc HealthCheck, name string) {
	registerHealthCheck(&optionalChecks, hc, name)
}

func registerHealthCheck(checks *sync.Map, hc HealthCheck, name string) {
	// check both lists, because
	if _, inReq := requiredChecks.Load(name); inReq {
		log.Warnf("tried to register health check with name %q twice", name)
		return
	}
	if _, inOpt := optionalChecks.Load(name); inOpt {
		log.Warnf("tried to register health check with name %q twice", name)
		return
	}
	if initHC, ok := hc.(Initialisable); ok {
		if err := initHC.Init(); err != nil {
			log.Warnf("error initialising health check %q: %s", name, err)
			initErrors.Store(name, &ConnectionState{
				lastCheck: time.Now(),
				result: HealthCheckResult{
					State: Err,
					Msg:   err.Error(),
				},
			})
		}
	}
	// save the length of the longest health check name, for the width of the column in /health/check
	if len(name) > longestCheckName {
		longestCheckName = len(name)
	}
	checks.Store(name, hc)
}

// HealthHandler returns the health endpoint for transactional processing. This Handler only checks
// the required health checks and returns ERR and 503 or OK and 200.
func HealthHandler() http.Handler {
	return &healthHandler{}
}

// ReadableHealthHandler returns the health endpoint with all details about service health. This handler checks
// all health checks. The response body contains two tables (for required and optional health checks)
// with the detailed results of the health checks.
func ReadableHealthHandler() http.Handler {
	return &readableHealthHandler{}
}
