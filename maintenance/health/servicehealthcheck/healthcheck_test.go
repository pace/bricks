// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.

package servicehealthcheck

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerHealthCheck(t *testing.T) {
	handler := HealthHandler()

	testCases := []struct {
		title   string
		check   *mockHealthCheck
		expCode int
		expBody string
	}{
		{
			title:   "Ok",
			check:   &mockHealthCheck{name: "TestOk"},
			expCode: http.StatusOK,
			expBody: "OK",
		},
		{
			title:   "Warn",
			check:   &mockHealthCheck{name: "TestWarn", healthCheckWarn: true},
			expCode: http.StatusOK,
			expBody: "OK",
		},
		{
			title:   "Init Error",
			check:   &mockHealthCheck{name: "TestHandlerInitErr", initErr: true},
			expCode: http.StatusServiceUnavailable,
			expBody: "ERR: 1 errors and 0 warnings",
		},
		{
			title:   "Error",
			check:   &mockHealthCheck{name: "TestHandlerHealthCheckErr", healthCheckErr: true},
			expCode: http.StatusServiceUnavailable,
			expBody: "ERR: 1 errors and 0 warnings",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			resetHealthChecks()
			// set warmup for unit testing explicitely to 0
			RegisterHealthCheck(tc.check.name, tc.check, UseWarmup(0))
			testRequest(t, handler, tc.expCode, expBody(tc.expBody))
		})
	}
}

func TestInitErrorRetryAndCaching(t *testing.T) {
	handler := HealthHandler()
	resetHealthChecks()

	bgInterval := time.Second
	{
		// Create Check with initErr
		hc := &mockHealthCheck{
			initErr:         true,
			healthCheckErr:  false,
			healthCheckWarn: false,
			name:            "initErr",
		}

		RegisterHealthCheck(hc.name, hc,
			UseInterval(bgInterval),
			UseInitErrResultTTL(time.Hour), // Big caching ttl of the init err result
		)
		testRequest(t, handler, http.StatusServiceUnavailable, expBody("ERR: 1 errors and 0 warnings"))

	}

	{
		hc := &mockHealthCheck{
			initErr:         true,
			healthCheckErr:  false,
			healthCheckWarn: false,
			name:            "initErr",
		}
		// No init err, but expect err because of cache
		hc.initErr = false
		waitForBackgroundCheck(bgInterval)
		testRequest(t, handler, http.StatusServiceUnavailable, expBody("ERR: 1 errors and 0 warnings"))
	}

	resetHealthChecks()

	{
		hc := &mockHealthCheck{
			initErr:         true,
			healthCheckErr:  false,
			healthCheckWarn: false,
			name:            "initErr",
		}
		// Expect err
		hc.initErr = true
		RegisterHealthCheck(hc.name, hc,
			UseInterval(bgInterval),
			UseInitErrResultTTL(0), // No caching of the init err results
		)
		waitForBackgroundCheck(bgInterval)
		testRequest(t, handler, http.StatusServiceUnavailable, expBody("ERR: 1 errors and 0 warnings"))
	}

	resetHealthChecks()

	{
		hc := &mockHealthCheck{
			initErr:         true,
			healthCheckErr:  false,
			healthCheckWarn: false,
			name:            "initErr",
		}

		// Remove init err, no caching, expect OK
		hc.initErr = false
		waitForBackgroundCheck(bgInterval)
		testRequest(t, handler, http.StatusOK, expBody("OK"))
	}
	resetHealthChecks()
}

func TestHandlerHealthCheckOptional(t *testing.T) {
	checkOpt := &mockHealthCheck{name: "TestHandlerHealthCheckErr", healthCheckErr: true}
	checkReq := &mockHealthCheck{name: "TestOk"}
	resetHealthChecks()

	RegisterHealthCheck(checkReq.name, checkReq)
	RegisterOptionalHealthCheck(checkOpt, checkOpt.name)

	testRequest(t, HealthHandler(), http.StatusOK, expBody("OK"))
}

// used in testRequest to customise the response body check
type resBodyComparer func(t *testing.T, resBody []byte)

// expBody will expect the response body to equal to the passed expected body.
func expBody(expBody string) resBodyComparer {
	return func(t *testing.T, resBody []byte) {
		t.Helper()
		require.Equal(t, expBody, string(resBody))
	}
}

// execute a test request on the given handler
// and assert the expected response code and body.
func testRequest(t *testing.T, handler http.Handler, expCode int, expBody resBodyComparer) {
	t.Helper()

	// Before making request, wait for the background check to execute at least once
	waitForBackgroundCheck()

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, nil)
	resp := rec.Result()
	assert.Equal(t, expCode, resp.StatusCode)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	if expBody != nil {
		expBody(t, data)
	}
}

// wait a bit + optional/additional delay,
// to allow the background check to kick in.
func waitForBackgroundCheck(additionalWait ...time.Duration) {
	t := 10 * time.Millisecond
	if len(additionalWait) > 0 {
		t += additionalWait[0]
	}
	time.Sleep(t)
}

// remove all previous health checks
func resetHealthChecks() {
	requiredChecks = sync.Map{}
	optionalChecks = sync.Map{}
}
