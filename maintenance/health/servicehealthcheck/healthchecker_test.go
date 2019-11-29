// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/18 by Charlotte Pröller

package servicehealthcheck

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/log"
)

type testHealthChecker struct {
	initErr        bool
	healthCheckErr bool
	name           string
}

func (t *testHealthChecker) Name() string {
	return t.name
}

func (t *testHealthChecker) Init() error {
	if t.initErr {
		return errors.New("initError")
	}
	return nil
}

func (t *testHealthChecker) HealthCheck() (bool, error) {
	if t.healthCheckErr {
		return false, errors.New("healtherror")
	}
	return true, nil
}

func setupOneHealthCheck(t *testHealthChecker) *http.Response {
	checks = sync.Map{}
	initErrors = sync.Map{}
	rec := httptest.NewRecorder()
	RegisterHealthCheck(t, t.Name())
	req := httptest.NewRequest("GET", "/health/"+t.Name(), nil)
	Handler().ServeHTTP(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	return resp
}

func TestHandlerOK(t *testing.T) {
	resp := setupOneHealthCheck(&testHealthChecker{name: "test"})
	if resp.StatusCode != 200 {
		t.Errorf("Expected /health to respond with 200, got: %d", resp.StatusCode)
	}
	helperCheckResponse(t, "OK\n", resp)
}

func TestHandlerInitErr(t *testing.T) {
	resp := setupOneHealthCheck(&testHealthChecker{name: "TestHandlerInitErr", initErr: true})
	if resp.StatusCode != 503 {
		t.Errorf("Expected /health to respond with 503, got: %d", resp.StatusCode)
	}
	helperCheckResponse(t, "ERR", resp)
}

func TestHandlerHealthCheckErr(t *testing.T) {
	resp := setupOneHealthCheck(&testHealthChecker{name: "TestHandlerHealthCheckErr", healthCheckErr: true})
	if resp.StatusCode != 503 {
		t.Errorf("Expected /health to respond with 503, got: %d", resp.StatusCode)
	}
	helperCheckResponse(t, "ERR", resp)

}

func TestHealthCheckAllWithErrors(t *testing.T) {

	hcs := []*testHealthChecker{
		{name: "TestHandlerHealthCheckErr", healthCheckErr: true},
		{name: "TestHandlerInitErr", initErr: true},
		{name: "test"},
	}
	expected := []string{
		"TestHandlerHealthCheckErr :ERR",
		"TestHandlerInitErr :ERR",
		"test :OK",
	}
	checkAllTestHcs(hcs, t, expected, 503)
}
func TestHealthCheckAllSuccess(t *testing.T) {
	hcs := []*testHealthChecker{
		{name: "test2"},
		{name: "test3"},
		{name: "test"},
	}
	expected := []string{
		"test3 :OK",
		"test2 :OK",
		"test :OK",
	}
	checkAllTestHcs(hcs, t, expected, 200)
}

func checkAllTestHcs(hcs []*testHealthChecker, t *testing.T, expected []string, statusCode int) {
	checks = sync.Map{}
	initErrors = sync.Map{}
	rec := httptest.NewRecorder()
	for _, hc := range hcs {
		RegisterHealthCheck(hc, hc.Name())
	}
	req := httptest.NewRequest("GET", "/health/all", nil)
	Handler().ServeHTTP(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	if resp.StatusCode != statusCode {
		t.Errorf("Expected /health to respond with %d, got: %d", statusCode, resp.StatusCode)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	checkRes := strings.Split(string(data), "\n")
	if exp := len(checkRes) - 1; exp != len(expected) {
		t.Errorf("Expected %v HealthCheck results, but got %v", len(expected), len(hcs)-1)
	}
	// Check of all healthCheck results are present and correct
	for exp := range expected {
		found := false
		for res := range checkRes {
			if res == exp {
				found = true
			}
		}
		if !found {
			t.Errorf("Expected health to return %q, got: %q", exp, string(data))
		}
	}
}

func helperCheckResponse(t *testing.T, expected string, resp *http.Response) {
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if string(data[:]) != expected {
		t.Errorf("Expected health to return %q, got: %q", expected,
			string(data[:]))
	}
}
