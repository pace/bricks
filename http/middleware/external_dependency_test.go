// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_ExternalDependency_Middleare(t *testing.T) {
	AddExternalDependency(context.TODO(), "test", time.Second) // should not fail
	t.Run("empty set", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		h := ExternalDependency(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		h.ServeHTTP(rec, req)
		assert.Nil(t, rec.Result().Header[ExternalDependencyHeaderName])
	})
	t.Run("one dependency set", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		h := ExternalDependency(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			AddExternalDependency(r.Context(), "test", time.Second)
			w.Write(nil) // nolint: errcheck
		}))
		h.ServeHTTP(rec, req)
		assert.Equal(t, rec.Result().Header[ExternalDependencyHeaderName][0], "test:1000")
	})
}

func Test_ExternalDependencyContext_String(t *testing.T) {
	var edc ExternalDependencyContext

	// empty
	assert.Empty(t, edc.String())

	// one dependency
	edc.AddDependency("test1", time.Millisecond)
	assert.EqualValues(t, "test1:1", edc.String())

	// multiple dependencies
	edc.AddDependency("test2", time.Nanosecond)
	assert.EqualValues(t, "test1:1,test2:0", edc.String())

	edc.AddDependency("test3", time.Second)
	assert.EqualValues(t, "test1:1,test2:0,test3:1000", edc.String())

	edc.AddDependency("test4", time.Millisecond*123)
	assert.EqualValues(t, "test1:1,test2:0,test3:1000,test4:123", edc.String())
}

func Test_ExternalDependencyContext_Parse(t *testing.T) {
	var edc ExternalDependencyContext

	// empty
	assert.Empty(t, edc.String())

	// one dependency
	edc.Parse("test1:1")
	assert.EqualValues(t, "test1:1", edc.String())

	// ignore invalid lines
	edc.Parse("error")
	assert.EqualValues(t, "test1:1", edc.String())

	// multiple dependencies
	edc.Parse("test2:0")
	assert.EqualValues(t, "test1:1,test2:0", edc.String())

	edc.Parse("test3:1000,test4:123")
	assert.EqualValues(t, "test1:1,test2:0,test3:1000,test4:123", edc.String())
}
