// Copyright © 2021 by PACE Telematics GmbH. All rights reserved.
// Created at 2021/04/27 by Alessandro Miceli

package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ResponseClientID_Middleare(t *testing.T) {
	AddResponseClientID(context.TODO(), "test") // should not fail
	t.Run("empty set", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		h := ResponseClientID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		}))
		h.ServeHTTP(rec, req)
		assert.Nil(t, rec.Result().Header[http.CanonicalHeaderKey(ClientIDHeaderName)])
	})
	t.Run("one client set", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		h := ResponseClientID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			AddResponseClientID(r.Context(), "test")
			w.Write(nil) // nolint: errcheck
		}))
		h.ServeHTTP(rec, req)
		assert.Equal(t, rec.Result().Header[http.CanonicalHeaderKey(ClientIDHeaderName)][0], "test")
	})
}

func Test_ResponseClientIDContext_String(t *testing.T) {
	var rcc ResponseClientIDContext

	// empty
	assert.Empty(t, rcc.String())

	// one dependency
	rcc.AddResponseClientID("test1")
	assert.EqualValues(t, "test1", rcc.String())

	// multiple dependencies
	rcc.AddResponseClientID("test2")
	assert.EqualValues(t, "test1,test2", rcc.String())

	rcc.AddResponseClientID("test3")
	assert.EqualValues(t, "test1,test2,test3", rcc.String())

	rcc.AddResponseClientID("test4")
	assert.EqualValues(t, "test1,test2,test3,test4", rcc.String())
}
