// Copyright © 2018–2026 by PACE Mobility GmbH. All rights reserved.

package metric

import (
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)

	Handler().ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Failed to respond with prometheus metrics: %v", resp.StatusCode)
	}
}
