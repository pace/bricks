// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/18 by Charlotte Pröller

package servicehealthcheck

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"

	//"strings"
	"sync"
	"testing"

	"github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/log"
)

type testHealthChecker struct {
	initErr         bool
	healthCheckErr  bool
	healthCheckWarn bool
	name            string
}

func (t *testHealthChecker) Init() error {
	if t.initErr {
		return errors.New("initError")
	}
	return nil
}

func (t *testHealthChecker) HealthCheck() HealthCheckResult {
	if t.healthCheckErr {
		return HealthCheckResult{Err, "healthCheckErr"}
	}
	return HealthCheckResult{Ok, ""}
}

func TestHandlerHealthCheck(t *testing.T) {
	testcases := []struct {
		title   string
		check   *testHealthChecker
		expCode int
		expBody string
	}{
		{
			title:   "Test HealthCheck Error",
			check:   &testHealthChecker{name: "TestHandlerHealthCheckErr", healthCheckErr: true},
			expCode: http.StatusServiceUnavailable,
			expBody: "ERR",
		},
		{
			title:   "Test HealthCheck init Error",
			check:   &testHealthChecker{name: "TestHandlerInitErr", initErr: true},
			expCode: http.StatusServiceUnavailable,
			expBody: "ERR",
		},
		{
			title:   "Test HealthCheck init and check Error",
			check:   &testHealthChecker{name: "TestHandlerInitErrHealthErr", initErr: true, healthCheckErr: true},
			expCode: http.StatusServiceUnavailable,
			expBody: "ERR",
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
	for _, tc := range testcases {
		t.Run(tc.title, func(t *testing.T) {
			//remove all previous health checks
			requiredChecks = sync.Map{}
			optionalChecks = sync.Map{}
			initErrors = sync.Map{}
			rec := httptest.NewRecorder()
			RegisterHealthCheck(tc.check, tc.check.name)
			req := httptest.NewRequest("GET", "/health/", nil)
			HealthHandler().ServeHTTP(rec, req)
			resp := rec.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tc.expCode {
				t.Errorf("Expected /health to respond with status code %d, got: %d", tc.expCode, resp.StatusCode)
			}
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			if string(data) != tc.expBody {
				t.Errorf("Expected /health to respond with body %q, got: %q", tc.expBody, string(data))
			}
		})
	}
}

func TestHandlerHealthCheckOptional(t *testing.T) {
	checkOpt := &testHealthChecker{name: "TestHandlerHealthCheckErr", healthCheckErr: true}
	checkReq := &testHealthChecker{name: "TestOk"}
	//remove all previous health checks
	requiredChecks = sync.Map{}
	optionalChecks = sync.Map{}
	initErrors = sync.Map{}

	rec := httptest.NewRecorder()
	RegisterHealthCheck(checkReq, checkReq.name)
	RegisterOptionalHealthCheck(checkOpt, checkOpt.name)

	req := httptest.NewRequest("GET", "/health/", nil)
	HealthHandler().ServeHTTP(rec, req)
	resp := rec.Result()

	if exp, got := http.StatusOK, resp.StatusCode; exp != got {
		t.Errorf("Expected /health to respond with status code %d, got: %d", exp, got)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if exp, got := "OK", string(data); exp != got {
		t.Errorf("Expected /health to respond with body %q, got: %q", exp, got)
	}
}

func TestHandlerReadableHealthCheck(t *testing.T) {
	longName := &testHealthChecker{name: "veryveryveryveryveryveryveryveryveryveryveryverylongname"}
	warn := &testHealthChecker{name: "WithWarning", healthCheckWarn: true}
	err := &testHealthChecker{name: "WithErr", healthCheckErr: true}
	initErr := &testHealthChecker{name: "WithInitErr", initErr: true}

	testcases := [] struct {
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
			expReq: [] string{
				"WithWarning                                                OK    ",
				"WithErr                                                    ERR   healthCheckErr",
				"WithInitErr                                                ERR   initError",
				"veryveryveryveryveryveryveryveryveryveryveryverylongname   OK    "},
		},
		{
			title:   "Test health check readable all optional",
			req:     []*testHealthChecker{},
			opt:     []*testHealthChecker{longName, warn, err, initErr},
			expCode: http.StatusServiceUnavailable,
			expOpt: [] string{
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
			expReq:  [] string{"veryveryveryveryveryveryveryveryveryveryveryverylongname   OK    "},
			expOpt:  [] string{"WithWarning                                                OK    "},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.title, func(t *testing.T) {
			//remove all previous health checks
			requiredChecks = sync.Map{}
			optionalChecks = sync.Map{}
			initErrors = sync.Map{}

			rec := httptest.NewRecorder()
			for _, hc := range tc.req {
				RegisterHealthCheck(hc, hc.name)
			}
			for _, hc := range tc.opt {
				RegisterOptionalHealthCheck(hc, hc.name)
			}
			req := httptest.NewRequest("GET", "/health/check", nil)
			ReadableHealthHandler().ServeHTTP(rec, req)
			resp := rec.Result()

			if resp.StatusCode != tc.expCode {
				t.Errorf("Expected /health to respond with status code %d, got: %d", tc.expCode, resp.StatusCode)
			}

			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}

			// health check results are not ordered because ranging over a map is randomized,
			// so the result is splitted into slices of health check results
			results := strings.Split(string(data), "Optional Services: \n")
			reqRes := strings.Split(strings.Split(results[0], "Required Services: \n")[1], "\n")
			optRes := strings.Split(results[1], "\n")
			testListHealthChecks(tc.expReq, reqRes, t, string(data))
			testListHealthChecks(tc.expOpt, optRes, t, string(data))
		})
	}
}

func testListHealthChecks(expected []string, checkResult []string, t *testing.T, body string) {
	// checkResult contains a empty string because of the splitting
	if expLength := len(checkResult) - 1; expLength != len(expected) {
		t.Errorf("Expected to have %d health check results, got: %d in body %q", expLength, len(checkResult), body)
	}
	// health check results are not ordered because ranging over a map is randomized,
	// so the rows needs to be sorted for easy comprehension
	sort.Strings(expected)
	sort.Strings(checkResult)

	for i := range expected {
		if expected[i] != checkResult[i+1] {
			t.Errorf("Expected health/check to return %q, got: %q", expected[i], checkResult[i+1])
		}
	}
}
