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
	"github.com/pace/bricks/maintenance/health"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health/liveness", nil)

	Router().ServeHTTP(rec, req)

	resp := rec.Result()
	require.Equal(t, 200, resp.StatusCode)

	data, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "OK\n", string(data))
}

func TestHealthRoutes(t *testing.T) {
	tCs := []struct {
		route          string
		expectedResult string
		title          string
	}{{
		route:          "/health",
		expectedResult: "OK",
		title:          "Route Health",
	}, {
		route:          "/health/check",
		expectedResult: "Required Services: \nOptional Services: \n",
		title:          "Route Health detailed",
	}, {
		route:          "/health/readiness",
		expectedResult: "Ready",
		title:          "Route readiness",
	}, {
		route:          "/health/liveness",
		expectedResult: "OK\n",
		title:          "route liveness",
	}}
	health.SetCustomReadinessCheck(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprint(w, "Ready")
		require.NoError(t, err)
	})
	for _, tC := range tCs {
		t.Run(tC.title, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", tC.route, nil)

			Router().ServeHTTP(rec, req)

			resp := rec.Result()
			data, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, tC.expectedResult, string(data))
		})
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

	require.Equal(t, 501, resp.StatusCode, "Expected /foo/bar to respond with 501")

	var e struct {
		List runtime.Errors `json:"errors"`
	}

	err := json.NewDecoder(resp.Body).Decode(&e)
	require.NoError(t, err)
	require.NotEmptyf(t, e.List[0].ID, "Expected first error to contain request ID, got: %#v", e.List[0])

}
