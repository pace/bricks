// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/21 by Vincent Landgraf

package tracing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opentracing/opentracing-go"

	"github.com/gorilla/mux"
)

func TestHandlerIgnore(t *testing.T) {
	r := mux.NewRouter()
	r.Use(Handler("/test"))
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(rec, req)
}

func TestHandler(t *testing.T) {
	r := mux.NewRouter()
	r.Use(Handler())
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/foo", nil)

	r.ServeHTTP(rec, req)
}

func TestRequest(t *testing.T) {
	r := Request(httptest.NewRequest("GET", "/test", nil))
	// check that header is empty
	if len(r.Header["Uber-Trace-Id"]) != 0 {
		t.Errorf("expected no tracing id but got one")
	}

	r = httptest.NewRequest("GET", "/test", nil)
	_, ctx := opentracing.StartSpanFromContext(context.Background(), "foo")
	r = Request(r.WithContext(ctx))
	if len(r.Header["Uber-Trace-Id"]) != 1 {
		t.Errorf("expected one tracing id but got none (JAEGER_SERVICE_NAME not in env?)")
	}
}
