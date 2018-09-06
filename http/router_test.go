// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/05 by Vincent Landgraf

package http

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"lab.jamit.de/pace/go-microservice/maintenance/log"
)

func TestHealthHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)

	Router().ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected /health to respond with 200, got: %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if string(data[:]) != "OK\n" {
		t.Errorf("Expected health to return OK, got: %q", string(data[:]))
	}
	if resp.Header.Get("Request-Id") == "" {
		t.Errorf("Expected response to contain Request-Id, got: %#v", resp.Header)
	}
}

func TestCustomRoutes(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/foo/bar", nil)

	// example of a service foo exposing api bar
	fooRouter := mux.NewRouter()
	fooRouter.HandleFunc("/foo/bar", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}).Methods("GET")

	r := Router()
	// service routers will be mounted like this
	r.PathPrefix("/foo/").Handler(fooRouter)

	r.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != 501 {
		t.Errorf("Expected /foo/bar to respond with 501, got: %d", resp.StatusCode)
	}
}
