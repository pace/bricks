// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package tracing

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pace/bricks/maintenance/util"
)

func TestHandlerIgnore(t *testing.T) {
	r := mux.NewRouter()
	r.Use(util.NewIgnorePrefixMiddleware(Handler(), "/test"))
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// This test does not tests if any prefix is ignored
	r.ServeHTTP(rec, req)
}

func TestHandler(t *testing.T) {
	r := mux.NewRouter()
	r.Use(Handler())
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	r.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() {
		err := resp.Body.Close()
		assert.NoError(t, err)
	}()

	// This test does not tests the tracing
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
