// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package health

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pace/bricks/maintenance/log"
	"github.com/stretchr/testify/require"
)

func TestHandlerLiveness(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health/liveness", nil)

	HandlerLiveness().ServeHTTP(rec, req)

	checkResult(rec, 200, "OK\n", t)
}

func TestHandlerReadiness(t *testing.T) {
	// check the default
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health/readiness", nil)
	HandlerReadiness().ServeHTTP(rec, req)

	// check another readiness check
	checkResult(rec, 200, "OK\n", t)
	rec = httptest.NewRecorder()
	SetCustomReadinessCheck(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte("Err\n")); err != nil {
			log.Warnf("could not write output: %s", err)
		}
	})
	HandlerReadiness().ServeHTTP(rec, req)
	checkResult(rec, 404, "Err\n", t)
}

func checkResult(rec *httptest.ResponseRecorder, expCode int, expBody string, t *testing.T) {
	resp := rec.Result()
	defer resp.Body.Close()
	if resp.StatusCode != expCode {
		t.Errorf("Expected /health to respond with %d, got: %d", expCode, resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, expBody, string(data))

}
