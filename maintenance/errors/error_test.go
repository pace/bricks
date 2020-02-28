// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/20 by Vincent Landgraf

package errors

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/pace/bricks/http/transport"
	"github.com/pace/bricks/maintenance/errors/raven"
	"github.com/pace/bricks/maintenance/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func Test_createBreadcrumb(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]interface{}
		want    *raven.Breadcrumb
		wantErr bool
	}{
		{
			name: "standard_error",
			data: map[string]interface{}{
				"level":   "error",
				"message": "this is an error message",
				"time":    "2020-02-27T10:19:28+01:00",
				"req_id":  "bpboj6bipt34r4teo7g0",
			},
			want: &raven.Breadcrumb{
				Level:     "error",
				Message:   "this is an error message",
				Timestamp: 1582795168,
				Data:      map[string]interface{}{},
			},
		},
		{
			name: "http",
			data: map[string]interface{}{
				"level":           "debug",
				"time":            "2020-02-27T10:19:28+01:00",
				"sentry:category": "http",
				"sentry:type":     "http",
				"message":         "HTTPS GET www.pace.car",
				"method":          "GET",
				"attempt":         1,
				"status_code":     200,
				"duration":        227.717783,
				"url":             "https://www.pace.car/",
				"req_id":          "bpboj6bipt34r4teo7g0",
			},
			want: &raven.Breadcrumb{
				Category:  "http",
				Level:     "debug",
				Message:   "HTTPS GET www.pace.car",
				Timestamp: 1582795168,
				Type:      "http",
				Data: map[string]interface{}{
					"method":      "GET",
					"attempt":     1,
					"status_code": 200,
					"duration":    227.717783,
					"url":         "https://www.pace.car/",
				},
			},
		},
		{
			name: "panic_level",
			data: map[string]interface{}{
				"level":   "panic",
				"message": "this is a panic message",
				"time":    "2020-02-27T10:19:28+01:00",
			},
			want: &raven.Breadcrumb{
				Level:     "fatal",
				Type:      "error",
				Message:   "this is a panic message",
				Timestamp: 1582795168,
				Data:      map[string]interface{}{},
			},
		},
		{
			name: "custom_category",
			data: map[string]interface{}{
				"level":           "info",
				"message":         "this is an error message",
				"sentry:category": "redis",
				"sentry:type":     "error",
				"time":            "2020-02-27T10:19:28+01:00",
				"req_id":          "bpboj6bipt34r4teo7g0",
			},
			want: &raven.Breadcrumb{
				Category:  "redis",
				Level:     "info",
				Timestamp: 1582795168,
				Message:   "this is an error message",
				Type:      "error",
				Data:      map[string]interface{}{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createBreadcrumb(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("createBreadcrumb() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, got, "createBreadcrumb() = %v, want %v", got, tt.want)
		})
	}
}

// TestHandlerWithLogSink tests whether the panic recover handler
// still works and the corresponding logs reach the integrated log.Sink
// which should be passed to all subsequent requests and handler.
func TestHandlerWithLogSink(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	var sinkCtx context.Context

	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		require.NotEqual(t, "", log.RequestID(r), "request should have request id")
		sinkCtx = r.Context()

		client := &http.Client{
			Transport: transport.NewDefaultTransportChain(),
		}

		r0, err := http.NewRequest("GET", "https://www.pace.car/de", nil)
		assert.NoError(t, err, `failed creating request to "/succeed"`)

		r0 = r0.WithContext(r.Context())
		_, err = client.Do(r0)
		assert.NoError(t, err, `request to "/succeed" should not error`)

		r1, err := http.NewRequest("GET", "http://localhost/fail", nil)
		assert.NoError(t, err, `failed creating request to "/fail"`)

		r1 = r1.WithContext(r.Context())
		_, err = client.Do(r1)
		assert.Error(t, err, `request to "/fail" should error`)

		log.Req(r).Info().
			Str("sentry:category", "redis").
			Str("sentry:type", "error").
			Msg("this is a test message for the sink")

		panic("Sink Test Error, IGNORE")
	})
	log.Handler()(Handler()(mux)).ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	sink, ok := log.SinkFromContext(sinkCtx)
	assert.True(t, ok, "failed getting sink")

	var logLines []json.RawMessage
	assert.NoError(t, json.Unmarshal(sink.ToJSON(), &logLines), "failed extracting logs from sink")

	assert.Contains(t, string(logLines[0]), "https://www.pace.car/de", "missing log line")
	assert.Contains(t, string(logLines[1]), "http://localhost/fail", "missing log line")

	assert.Contains(t, string(logLines[2]), "sentry:category", "missing log line")
	assert.Contains(t, string(logLines[2]), "redis", "missing log line")
	assert.Contains(t, string(logLines[2]), "sentry:type", "missing log line")
	assert.Contains(t, string(logLines[2]), "error", "missing log line")
	assert.Contains(t, string(logLines[2]), "this is a test message for the sink", "missing log line")

	assert.Contains(t, string(logLines[3]), "Sink Test Error, IGNORE", "missing log line")

	require.Equal(t, 500, resp.StatusCode, "wrong status code")
}
