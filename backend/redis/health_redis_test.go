// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.

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
	req := httptest.NewRequest(http.MethodGet, "/health/check", nil)
	r.ServeHTTP(rec, req)

	return rec.Result()
}

// TestIntegrationHealthCheck tests if redis health check ist working like expected.
func TestIntegrationHealthCheck(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	time.Sleep(time.Second)

	resp := setup()

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
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
