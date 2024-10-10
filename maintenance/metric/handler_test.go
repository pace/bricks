// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package metric

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)

	Handler().ServeHTTP(rec, req)

	resp := rec.Result()

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Failed to respond with prometheus metrics: %v", resp.StatusCode)
	}
}
