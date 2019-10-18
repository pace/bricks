// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/12 by Vincent Landgraf

package postgres

import (
	"github.com/gorilla/mux"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
	"github.com/pace/bricks/maintenance/log"
	"io/ioutil"
	"net/http"

	"net/http/httptest"
	"testing"
)

var resp *http.Response

func setup(t *postgresHealthCheck) {
	r := mux.NewRouter()
	rec := httptest.NewRecorder()
	servicehealthcheck.InitialiseHealthChecker(r)
	servicehealthcheck.RegisterHealthCheck(t)
	req := httptest.NewRequest("GET", "/health/"+t.Name(), nil)
	r.ServeHTTP(rec, req)
	resp = rec.Result()
	defer resp.Body.Close()
}

func TestIntegrationHealthCheck(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	setup(&postgresHealthCheck{})
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
