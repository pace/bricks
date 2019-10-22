// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/18 Charlotte Pröller

package servicehealthcheck

import (
	"github.com/gorilla/mux"
	"github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/log"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testHealthChecker struct {
	initErr        bool
	healthCheckErr bool
	name           string
}

func (t *testHealthChecker) Name() string {
	return t.name
}

func (t *testHealthChecker) InitHealthCheck() error {
	if t.initErr {
		return errors.New("initError")
	} else {
		return nil
	}
}

func (t *testHealthChecker) HealthCheck() (bool, error) {
	if t.healthCheckErr {
		return false, errors.New("healtherror")
	} else {
		return true, nil
	}
}

func (t *testHealthChecker) CleanUp() error {
	return nil
}

var resp *http.Response

func setup(t *testHealthChecker) {
	r := mux.NewRouter()
	rec := httptest.NewRecorder()
	InitialiseHealthChecker(r)
	RegisterHealthCheck(t, t.Name())
	req := httptest.NewRequest("GET", "/health/"+t.Name(), nil)
	Handler().ServeHTTP(rec, req)
	resp = rec.Result()
	defer resp.Body.Close()

}

func TestHandlerOK(t *testing.T) {
	setup(&testHealthChecker{name: "test"})
	if resp.StatusCode != 200 {
		t.Errorf("Expected /health to respond with 200, got: %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if string(data[:]) != "OK\n" {
		t.Errorf("Expected health to return OK, got: %q", string(data[:]))
	}
}

func TestHandlerInitErr(t *testing.T) {
	setup(&testHealthChecker{name: "TestHandlerInitErr", initErr: true})
	if resp.StatusCode != 503 {
		t.Errorf("Expected /health to respond with 503, got: %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if string(data[:]) != "Error \n" {
		t.Errorf("Expected health to return Error \\n, got: %q", string(data[:]))
	}
}

func TestHandlerHealthCheckErr(t *testing.T) {
	setup(&testHealthChecker{name: "TestHandlerHealthCheckErr", healthCheckErr: true})
	if resp.StatusCode != 503 {
		t.Errorf("Expected /health to respond with 503, got: %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if string(data[:]) != "Error \n" {
		t.Errorf("Expected health to return Error \\n, got: %q", string(data[:]))
	}
}
