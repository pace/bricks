// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/18 by Charlotte Pröller

package servicehealthcheck

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testHealthChecker struct {
	initErr         bool
	healthCheckErr  bool
	healthCheckWarn bool
	name            string
}

func (t *testHealthChecker) Init(ctx context.Context) error {
	if t.initErr {
		return errors.New("initError")
	}
	return nil
}

func (t *testHealthChecker) HealthCheck(ctx context.Context) HealthCheckResult {
	if t.healthCheckErr {
		return HealthCheckResult{Err, "healthCheckErr"}
	}
	return HealthCheckResult{Ok, ""}
}

type testBackgroundHealthChecker struct {
	*testHealthChecker
	interval time.Duration
}

func (t *testBackgroundHealthChecker) Interval() time.Duration {
	return t.interval
}

func TestHandlerHealthCheck(t *testing.T) {
	testCases := []struct {
		title   string
		check   *testHealthChecker
		expCode int
		expBody string
	}{
		{
			title:   "Test HealthCheck Error",
			check:   &testHealthChecker{name: "TestHandlerHealthCheckErr", healthCheckErr: true},
			expCode: http.StatusServiceUnavailable,
			expBody: "ERR: 1 errors and 0 warnings",
		},
		{
			title:   "Test HealthCheck init Error",
			check:   &testHealthChecker{name: "TestHandlerInitErr", initErr: true},
			expCode: http.StatusServiceUnavailable,
			expBody: "ERR: 1 errors and 0 warnings",
		},
		{
			title:   "Test HealthCheck init and check Error",
			check:   &testHealthChecker{name: "TestHandlerInitErrHealthErr", initErr: true, healthCheckErr: true},
			expCode: http.StatusServiceUnavailable,
			expBody: "ERR: 1 errors and 0 warnings",
		},
		{
			title:   "Test HealthCheck Ok",
			check:   &testHealthChecker{name: "TestOk"},
			expCode: http.StatusOK,
			expBody: "OK",
		},
		{
			title:   "Test HealthCheck Warn",
			check:   &testHealthChecker{name: "TestWarn", healthCheckWarn: true},
			expCode: http.StatusOK,
			expBody: "OK",
		},
	}
	for _, tc := range testCases {
		// Test each healthcheck both in direct mode and background mode
		t.Run(tc.title, func(t *testing.T) {
			t.Run("direct", func(t *testing.T) {
				resetHealthChecks()
				var hc HealthCheck = tc.check
				RegisterHealthCheck(tc.check.name, hc)
				testRequest(t, tc.expCode, tc.expBody)
			})
			t.Run("background", func(t *testing.T) {
				resetHealthChecks()
				var hc HealthCheck = &testBackgroundHealthChecker{testHealthChecker: tc.check}
				RegisterHealthCheck(tc.check.name, hc)
				waitForBackgroundCheck()
				testRequest(t, tc.expCode, tc.expBody)
			})
		})
	}
}

// waitForBackgroundCheck waits a bit + optional/additional delay, to allow the background check to kick in.
func waitForBackgroundCheck(additionalWait ...time.Duration) {
	t := 10 * time.Millisecond
	if len(additionalWait) > 0 {
		t += additionalWait[0]
	}
	time.Sleep(t)
}

func TestInitErrorRetry(t *testing.T) {
	// No caching of the init results
	cfg.HealthCheckInitResultErrorTTL = 0

	resetHealthChecks()

	// Create Check with initErr
	checker := &testHealthChecker{
		initErr:         true,
		healthCheckErr:  false,
		healthCheckWarn: false,
		name:            "initRetry",
	}
	RegisterHealthCheck(checker.name, checker)

	testRequest(t, http.StatusServiceUnavailable, "ERR: 1 errors and 0 warnings")

	// remove initErr
	checker.initErr = false
	testRequest(t, http.StatusOK, "OK")
}

func TestInitErrorRetry_BackgroundCheck(t *testing.T) {
	// No caching of the init err results
	cfg.HealthCheckInitResultErrorTTL = 0

	resetHealthChecks()

	// Create Check with initErr
	checker := &testBackgroundHealthChecker{
		testHealthChecker: &testHealthChecker{
			initErr:         true,
			healthCheckErr:  false,
			healthCheckWarn: false,
			name:            "initRetry",
		},
		interval: time.Second,
	}
	RegisterHealthCheck(checker.name, checker)
	waitForBackgroundCheck()
	testRequest(t, http.StatusServiceUnavailable, "ERR: 1 errors and 0 warnings")

	// remove initErr
	checker.initErr = false
	waitForBackgroundCheck(checker.interval)
	testRequest(t, http.StatusOK, "OK")
}

func testRequest(t *testing.T, expCode int, expBody string) {
	t.Helper()

	req := httptest.NewRequest("GET", "/health/", nil)
	rec := httptest.NewRecorder()
	HealthHandler().ServeHTTP(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	assert.Equal(t, expCode, resp.StatusCode)
	data, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, expBody, string(data))
}

func resetHealthChecks() {
	//remove all previous health checks
	requiredChecks = sync.Map{}
	optionalChecks = sync.Map{}
	initErrors = sync.Map{}
}

func TestInitErrorCaching(t *testing.T) {
	cfg.HealthCheckInitResultErrorTTL = time.Hour
	hc := &testHealthChecker{
		initErr:         true,
		healthCheckErr:  false,
		healthCheckWarn: false,
		name:            "initErrorCaching",
	}

	resetHealthChecks()
	RegisterHealthCheck(hc.name, hc)

	testRequest(t, http.StatusServiceUnavailable, "ERR: 1 errors and 0 warnings")
	hc.initErr = false
	testRequest(t, http.StatusServiceUnavailable, "ERR: 1 errors and 0 warnings")

	cfg.HealthCheckInitResultErrorTTL = 0
	resetHealthChecks()
	hc.initErr = true
	RegisterHealthCheck(hc.name, hc)
	testRequest(t, http.StatusServiceUnavailable, "ERR: 1 errors and 0 warnings")
	hc.initErr = false
	testRequest(t, http.StatusOK, "OK")
}

func TestInitErrorCaching_BackgroundCheck(t *testing.T) {
	// Big caching ttl of the init err results
	cfg.HealthCheckInitResultErrorTTL = time.Hour
	// Create Check with initErr
	hc := &testBackgroundHealthChecker{
		testHealthChecker: &testHealthChecker{
			initErr:         true,
			healthCheckErr:  false,
			healthCheckWarn: false,
			name:            "initErrorCaching",
		},
		interval: time.Second,
	}
	resetHealthChecks()
	RegisterHealthCheck(hc.name, hc)
	waitForBackgroundCheck()
	testRequest(t, http.StatusServiceUnavailable, "ERR: 1 errors and 0 warnings")

	// No init err, but expect err because of cache
	hc.initErr = false
	waitForBackgroundCheck(hc.interval)
	testRequest(t, http.StatusServiceUnavailable, "ERR: 1 errors and 0 warnings")

	// No caching of the init err results
	cfg.HealthCheckInitResultErrorTTL = 0
	resetHealthChecks()
	// Expect err
	hc.initErr = true
	RegisterHealthCheck(hc.name, hc)
	waitForBackgroundCheck()
	testRequest(t, http.StatusServiceUnavailable, "ERR: 1 errors and 0 warnings")

	// Remove init err, expect OK
	hc.initErr = false
	waitForBackgroundCheck(hc.interval)
	testRequest(t, http.StatusOK, "OK")
}

func TestHandlerHealthCheckOptional(t *testing.T) {
	checkOpt := &testHealthChecker{name: "TestHandlerHealthCheckErr", healthCheckErr: true}
	checkReq := &testHealthChecker{name: "TestOk"}
	resetHealthChecks()

	RegisterHealthCheck(checkReq.name, checkReq)
	RegisterOptionalHealthCheck(checkOpt, checkOpt.name)

	testRequest(t, http.StatusOK, "OK")
}

func TestHandlerReadableHealthCheck(t *testing.T) {
	longName := &testHealthChecker{name: "veryveryveryveryveryveryveryveryveryveryveryverylongname"}
	warn := &testHealthChecker{name: "WithWarning", healthCheckWarn: true}
	err := &testHealthChecker{name: "WithErr", healthCheckErr: true}
	initErr := &testHealthChecker{name: "WithInitErr", initErr: true}

	testcases := []struct {
		title   string
		req     []*testHealthChecker
		opt     []*testHealthChecker
		expCode int
		expReq  []string
		expOpt  []string
	}{
		{
			title:   "Test health check readable all required",
			req:     []*testHealthChecker{longName, warn, err, initErr},
			opt:     []*testHealthChecker{},
			expCode: http.StatusServiceUnavailable,
			expReq: []string{
				"WithWarning                                                OK    ",
				"WithErr                                                    ERR   healthCheckErr",
				"WithInitErr                                                ERR   initError",
				"veryveryveryveryveryveryveryveryveryveryveryverylongname   OK    "},
		},
		{
			title:   "Test health check readable all optional",
			req:     []*testHealthChecker{},
			opt:     []*testHealthChecker{longName, warn, err, initErr},
			expCode: http.StatusOK,
			expOpt: []string{
				"WithWarning                                                OK    ",
				"WithErr                                                    ERR   healthCheckErr",
				"WithInitErr                                                ERR   initError",
				"veryveryveryveryveryveryveryveryveryveryveryverylongname   OK    "},
		},
		{
			title:   "Test health check readable ok, all duplicated and with warning",
			req:     []*testHealthChecker{longName, longName},
			opt:     []*testHealthChecker{longName, warn},
			expCode: http.StatusOK,
			expReq:  []string{"veryveryveryveryveryveryveryveryveryveryveryverylongname   OK    "},
			expOpt:  []string{"WithWarning                                                OK    "},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.title, func(t *testing.T) {
			resetHealthChecks()

			rec := httptest.NewRecorder()
			for _, hc := range tc.req {
				RegisterHealthCheck(hc.name, hc)
			}
			for _, hc := range tc.opt {
				RegisterOptionalHealthCheck(hc, hc.name)
			}
			req := httptest.NewRequest("GET", "/health/check", nil)
			ReadableHealthHandler().ServeHTTP(rec, req)
			resp := rec.Result()

			require.Equal(t, tc.expCode, resp.StatusCode)

			data, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			// health check results are not ordered because ranging over a map is randomized,
			// so the result is splitted into slices of health check results
			results := strings.Split(string(data), "Optional Services: \n")
			reqRes := strings.Split(strings.Split(results[0], "Required Services: \n")[1], "\n")
			optRes := strings.Split(results[1], "\n")
			testListHealthChecks(tc.expReq, reqRes, t)
			testListHealthChecks(tc.expOpt, optRes, t)
		})
	}
}

func TestHandlerJSONHealthCheck(t *testing.T) {
	warn := &testHealthChecker{name: "WithWarning", healthCheckWarn: true}
	err := &testHealthChecker{name: "WithErr", healthCheckErr: true}
	initErr := &testHealthChecker{name: "WithInitErr", initErr: true}

	testcases := []struct {
		title   string
		req     []*testHealthChecker
		opt     []*testHealthChecker
		expCode int
		expReq  jsonHealthHandler
	}{
		{
			title:   "Test health check json all required",
			req:     []*testHealthChecker{warn, err, initErr},
			opt:     []*testHealthChecker{},
			expCode: http.StatusServiceUnavailable,
			expReq: jsonHealthHandler{
				err.name: serviceStats{
					Status:   Err,
					Required: true,
					Error:    "healthCheckErr",
				},
				initErr.name: serviceStats{
					Status:   Err,
					Required: true,
					Error:    "initError",
				},
				warn.name: serviceStats{
					Status:   Ok,
					Required: true,
					Error:    "",
				},
			},
		},
		{
			title:   "Test health check json some required, some optional",
			req:     []*testHealthChecker{warn, err},
			opt:     []*testHealthChecker{initErr},
			expCode: http.StatusServiceUnavailable,
			expReq: jsonHealthHandler{
				err.name: serviceStats{
					Status:   Err,
					Required: true,
					Error:    "healthCheckErr",
				},
				initErr.name: serviceStats{
					Status:   Err,
					Required: false,
					Error:    "initError",
				},
				warn.name: serviceStats{
					Status:   Ok,
					Required: true,
					Error:    "",
				},
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.title, func(t *testing.T) {
			resetHealthChecks()

			rec := httptest.NewRecorder()
			for _, hc := range tc.req {
				RegisterHealthCheck(hc.name, hc)
			}
			for _, hc := range tc.opt {
				RegisterOptionalHealthCheck(hc, hc.name)
			}
			req := httptest.NewRequest("GET", "/health/check", nil)
			JSONHealthHandler().ServeHTTP(rec, req)
			resp := rec.Result()

			require.Equal(t, tc.expCode, resp.StatusCode)

			var data jsonHealthHandler
			err := json.NewDecoder(resp.Body).Decode(&data)
			require.NoError(t, err)

			for k, v := range tc.expReq {
				require.Equal(t, v.Status, data[k].Status)
				require.Equal(t, v.Required, data[k].Required)
				require.Equal(t, v.Error, data[k].Error)
			}
		})
	}
}

func testListHealthChecks(expected []string, checkResult []string, t *testing.T) {
	// checkResult contains a empty string because of the splitting
	require.Equal(t, len(expected), len(checkResult)-1, "The amount of health check results in the response body needs to be equal to the amount of expected health check results.")

	// health check results are not ordered because ranging over a map is randomized,
	// so the rows needs to be sorted for easy comprehension
	sort.Stable(sort.StringSlice(expected))
	sort.Stable(sort.StringSlice(checkResult))

	for i := range expected {
		require.Equal(t, expected[i], checkResult[i+1], "The entry of the health check table is wrong")
	}
}
