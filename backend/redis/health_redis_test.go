// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/29 by Charlotte Pröller

package redis

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	http2 "github.com/pace/bricks/http"
	"github.com/pace/bricks/maintenance/log"
)

func setup() *http.Response {
	r := http2.Router()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health/check", nil)
	r.ServeHTTP(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	return resp
}

// TestIntegrationHealthCheck tests if redis health check ist working like expected
func TestIntegrationHealthCheck(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	time.Sleep(time.Second)
	resp := setup()
	if resp.StatusCode != 200 {
		t.Errorf("Expected /health/check to respond with 200, got: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if !strings.Contains(string(data), "redis                  OK") {
		t.Errorf("Expected /health/check to return OK, got: %q", string(data[:]))
	}
}
