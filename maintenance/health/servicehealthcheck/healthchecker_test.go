// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/18 by Charlotte Pröller

package servicehealthcheck

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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

func setup(t *testHealthChecker) *http.Response {
	rec := httptest.NewRecorder()
	RegisterHealthCheck(t, t.Name())
	req := httptest.NewRequest("GET", "/health/"+t.Name(), nil)
	Handler().ServeHTTP(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	return resp
}

func TestHandlerOK(t *testing.T) {
	resp := setup(&testHealthChecker{name: "test"})
	if resp.StatusCode != 200 {
		t.Errorf("Expected /health to respond with 200, got: %d", resp.StatusCode)
	}
	helperCheckResponse(t, "OK\n", resp)
}

func TestHandlerInitErr(t *testing.T) {
	resp := setup(&testHealthChecker{name: "TestHandlerInitErr", initErr: true})
	if resp.StatusCode != 503 {
		t.Errorf("Expected /health to respond with 503, got: %d", resp.StatusCode)
	}
	helperCheckResponse(t, "ERR", resp)
}

func TestHandlerHealthCheckErr(t *testing.T) {
	resp := setup(&testHealthChecker{name: "TestHandlerHealthCheckErr", healthCheckErr: true})
	if resp.StatusCode != 503 {
		t.Errorf("Expected /health to respond with 503, got: %d", resp.StatusCode)
	}
	helperCheckResponse(t, "ERR", resp)

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
