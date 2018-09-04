// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/04 by Vincent Landgraf

package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCaptureStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}

	cap := &CaptureStatus{ResponseWriter: rec}

	handler(cap, req)

	resp := rec.Result()

	if resp.StatusCode != 204 {
		t.Errorf("Failed to return correct 204 response status, got: %v", resp.StatusCode)
	}
	if cap.StatusCode != 204 {
		t.Errorf("Failed to capture correct 204 response status, got: %v", cap.StatusCode)
	}
}
