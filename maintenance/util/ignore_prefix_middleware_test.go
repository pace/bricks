// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/12/17 by Charlotte Pröller

package util

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestMiddlewareWithBlacklist(t *testing.T) {

	// setup the router
	handler := handlerWithStatusCode(http.StatusOK)
	thisHandler := handlerWithStatusCode(http.StatusInternalServerError)
	prefix := "/test"
	r := mux.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return NewIgnorePrefixMiddleware(next, thisHandler, prefix)
	})
	r.HandleFunc(prefix+"/anything", handler.ServeHTTP)
	r.HandleFunc("/anything", handler.ServeHTTP)

	// test the middleware
	testCases := []struct {
		title              string
		path               string
		statusCodeExpected int
	}{
		{title: "ignore the request", path: prefix + "/anything", statusCodeExpected: http.StatusOK},
		{title: "don't ignore the request", path: "/anything", statusCodeExpected: http.StatusInternalServerError},
	}
	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", tc.path, nil)
			r.ServeHTTP(rec, req)
			resp := rec.Result()
			require.Equal(t, tc.statusCodeExpected, resp.StatusCode)

		})
	}
}

type testHandler struct {
	statusCode int
}

func (t *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(t.statusCode)
}

func handlerWithStatusCode(statusCode int) http.Handler {
	return &testHandler{statusCode: statusCode}
}
