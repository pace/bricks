// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/18 by Charlotte Pröller

package postgres

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	http2 "github.com/pace/bricks/http"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
	"github.com/pace/bricks/maintenance/log"
)

func setup(t *HealthCheck, name string) *http.Response {
	r := http2.Router()
	rec := httptest.NewRecorder()
	servicehealthcheck.RegisterHealthCheck(t, name)
	req := httptest.NewRequest("GET", "/health/"+name, nil)
	r.ServeHTTP(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	return resp
}

func TestIntegrationHealthCheck(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	resp := setup(&HealthCheck{
		Pool: ConnectionPool(),
	}, "postgres")
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
