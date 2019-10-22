// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/12 by Vincent Landgraf

package redis

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

func setup(t *redisHealthCheck, name string) {
	r := mux.NewRouter()
	rec := httptest.NewRecorder()
	servicehealthcheck.InitialiseHealthChecker(r)
	servicehealthcheck.RegisterHealthCheck(t, name)
	req := httptest.NewRequest("GET", "/health/"+name, nil)
	r.ServeHTTP(rec, req)
	resp = rec.Result()
	defer resp.Body.Close()
}

// TestIntegrationHealthCheck tests if redis health check ist working like expected
func TestIntegrationHealthCheck(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	setup(&redisHealthCheck{}, "redis")
	if resp.StatusCode != 200 {
		t.Errorf("Expected /health/redis to respond with 200, got: %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if string(data[:]) != "OK\n" {
		t.Errorf("Expected /health/redis to return OK, got: %q", string(data[:]))
	}

}
