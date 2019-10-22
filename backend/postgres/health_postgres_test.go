// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/18 Charlotte Pröller

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

func setup(t *PgHealthCheck, name string) {
	r := mux.NewRouter()
	rec := httptest.NewRecorder()
	servicehealthcheck.InitialiseHealthChecker(r)
	servicehealthcheck.RegisterHealthCheck(t, name)
	req := httptest.NewRequest("GET", "/health/"+name, nil)
	r.ServeHTTP(rec, req)
	resp = rec.Result()
	defer resp.Body.Close()
}

func TestIntegrationHealthCheck(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	setup(&PgHealthCheck{}, "postgres")
	if resp.StatusCode != 200 {
		t.Errorf("Expected /health/postgres to respond with 200, got: %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if string(data[:]) != "OK\n" {
		t.Errorf("Expected /health/postgres to return OK, got: %q", string(data[:]))
	}

}
