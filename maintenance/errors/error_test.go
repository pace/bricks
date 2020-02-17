// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/20 by Vincent Landgraf

package errors

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// Note: Tests that there is no panic if there are no
// sentry infos provided

func TestHandler(t *testing.T) {
	mux := mux.NewRouter()
	mux.Use(Handler())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		panic("fire")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	mux.ServeHTTP(rec, req)

	if rec.Code != 500 {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), `"error":"Error"`) {
		t.Errorf(`Expected "error":"Error", got %q`, rec.Body.String())
	}
}

func TestHandleWithCtx(t *testing.T) {
	func() {
		defer HandleWithCtx(context.Background(), "sample")
		panic("sample")
	}()
}

func TestNew(t *testing.T) {
	if New("Test").Error() != "Test" {
		t.Error("invalid implementation of errors.New")
	}
}

func TestWrapWithExtra(t *testing.T) {
	if WrapWithExtra(New("Test"), map[string]interface{}{}).Error() != "Test" {
		t.Error("invalid implementation of errors.WrapWithExtra")
	}
}

func TestStackTrace(t *testing.T) {
	e := sentryEvent{
		ctx:         context.Background(),
		handlerName: "TestStackTrace",
		r:           nil,
		level:       1,
		req:         nil,
	}
	pak := e.build()

	d, err := pak.JSON()
	assert.NoError(t, err)

	assert.NotContains(t, string(d), `"module":"testing"`)
	assert.NotContains(t, string(d), `"filename":"testing/testing.go"`)
}
