// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/18 Charlotte Pröller

package servicehealthcheck

import (
	"github.com/gorilla/mux"
	"github.com/pace/bricks/maintenance/errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func (t *testHealthChecker) InitHealthCheck() error {
	if t.initErr {
		return errors.New("initError")
	} else {
		return nil
	}
}

func (t *testHealthChecker) HealthCheck(currTime time.Time) (bool, error) {
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
	RegisterHealthCheck(t)
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
	if string(data[:]) != "Not OK:\ninitError" {
		t.Errorf("Expected health to return Not OK:\ninitError, got: %q", string(data[:]))
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
	if string(data[:]) != "Not OK:\nhealtherror" {
		t.Errorf("Expected health to return Not OK:\nhealtherror, got: %q", string(data[:]))
	}
}
