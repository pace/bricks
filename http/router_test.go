// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/05 by Vincent Landgraf

package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/pace/bricks/http/jsonapi/runtime"
	"github.com/pace/bricks/maintenance/log"
)

func TestHealthHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health/liveness", nil)

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
		runtime.WriteError(w, http.StatusNotImplemented, fmt.Errorf("Some error"))
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

	var e struct {
		List runtime.Errors `json:"errors"`
	}

	err := json.NewDecoder(resp.Body).Decode(&e)
	if err != nil {
		t.Fatal(err)
	}
	if e.List[0].ID == "" {
		t.Errorf("Expected first error to contain request ID, got: %#v", e.List[0])
	}
}
