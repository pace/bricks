// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package tracing

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pace/bricks/maintenance/util"
	"github.com/stretchr/testify/require"

	"github.com/gorilla/mux"
)

func TestHandlerIgnore(t *testing.T) {
	r := mux.NewRouter()
	r.Use(util.NewIgnorePrefixMiddleware(Handler(), "/test"))
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// This test does not tests if any prefix is ignored
	r.ServeHTTP(rec, req)
}

func TestHandler(t *testing.T) {
	r := mux.NewRouter()
	r.Use(Handler())
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	r.ServeHTTP(rec, req)

	// This test does not tests the tracing
	require.Equal(t, 200, rec.Result().StatusCode)
}
