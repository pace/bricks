// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package errors

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pace/bricks/http/transport"
	"github.com/pace/bricks/maintenance/log"
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
	req := httptest.NewRequest(http.MethodGet, "/", nil)

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
	if WrapWithExtra(New("Test"), map[string]any{}).Error() != "Test" {
		t.Error("invalid implementation of errors.WrapWithExtra")
	}
}

func Test_createBreadcrumb(t *testing.T) {
	tm, err := time.Parse(time.RFC3339, "2020-02-27T10:19:28+01:00")
	require.NoError(t, err)

	tests := []struct {
		name    string
		data    map[string]any
		want    *sentry.Breadcrumb
		wantErr bool
	}{
		{
			name: "standard_error",
			data: map[string]any{
				"level":   "error",
				"message": "this is an error message",
				"time":    "2020-02-27T10:19:28+01:00",
				"req_id":  "bpboj6bipt34r4teo7g0",
			},
			want: &sentry.Breadcrumb{
				Level:     "error",
				Message:   "this is an error message",
				Timestamp: tm,
				Data:      map[string]any{},
			},
		},
		{
			name: "http",
			data: map[string]any{
				"level":           "debug",
				"time":            "2020-02-27T10:19:28+01:00",
				"sentry:category": "http",
				"sentry:type":     "http",
				"message":         "HTTPS GET www.pace.car",
				"method":          http.MethodGet,
				"attempt":         1,
				"status_code":     http.StatusOK,
				"duration":        227.717783,
				"url":             "https://www.pace.car/",
				"req_id":          "bpboj6bipt34r4teo7g0",
			},
			want: &sentry.Breadcrumb{
				Category:  "http",
				Level:     "debug",
				Message:   "HTTPS GET www.pace.car",
				Timestamp: tm,
				Type:      "http",
				Data: map[string]any{
					"method":      http.MethodGet,
					"attempt":     1,
					"status_code": http.StatusOK,
					"duration":    227.717783,
					"url":         "https://www.pace.car/",
				},
			},
		},
		{
			name: "panic_level",
			data: map[string]any{
				"level":   "panic",
				"message": "this is a panic message",
				"time":    "2020-02-27T10:19:28+01:00",
			},
			want: &sentry.Breadcrumb{
				Level:     "fatal",
				Type:      "error",
				Message:   "this is a panic message",
				Timestamp: tm,
				Data:      map[string]any{},
			},
		},
		{
			name: "custom_category",
			data: map[string]any{
				"level":           "info",
				"message":         "this is an error message",
				"sentry:category": "redis",
				"sentry:type":     "error",
				"time":            "2020-02-27T10:19:28+01:00",
				"req_id":          "bpboj6bipt34r4teo7g0",
			},
			want: &sentry.Breadcrumb{
				Category:  "redis",
				Level:     "info",
				Timestamp: tm,
				Message:   "this is an error message",
				Type:      "error",
				Data:      map[string]any{},
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
	rec1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/test1", nil)

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/test2", nil)

	var (
		sink1Ctx context.Context
		sink2Ctx context.Context
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/test1", func(w http.ResponseWriter, r *http.Request) {
		sink1Ctx = r.Context()

		log.Ctx(r.Context()).Debug().Msg("ONLY FOR SINK1")
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/test2", func(w http.ResponseWriter, r *http.Request) {
		require.NotEqual(t, "", log.RequestID(r), "request should have request id")

		sink2Ctx = r.Context()

		client := &http.Client{
			Transport: transport.Chain(&transport.LoggingRoundTripper{}, &transport.DumpRoundTripper{}),
		}

		r0, err := http.NewRequest(http.MethodGet, "https://www.pace.car/de", nil)
		assert.NoError(t, err, `failed creating request to "/succeed"`)

		r0 = r0.WithContext(r.Context())

		resp, err := client.Do(r0)
		assert.NoError(t, err, `request to "/succeed" should not error`)

		defer func() {
			err := resp.Body.Close()
			assert.NoError(t, err)
		}()

		r1, err := http.NewRequest(http.MethodGet, "http://localhost/fail", nil)
		assert.NoError(t, err, `failed creating request to "/fail"`)

		r1 = r1.WithContext(r.Context())

		_, err = client.Do(r1) //nolint:bodyclose
		assert.Error(t, err, `request to "/fail" should error`)

		log.Req(r).Info().
			Str("sentry:category", "redis").
			Str("sentry:type", "error").
			Msg("ONLY FOR SINK2")

		panic("Sink2 Test Error, IGNORE")
	})

	handler := log.Handler()(Handler()(mux))

	handler.ServeHTTP(rec1, req1)

	resp1 := rec1.Result()
	require.Equal(t, http.StatusOK, resp1.StatusCode, "wrong status code")

	err := resp1.Body.Close()
	assert.NoError(t, err)

	handler.ServeHTTP(rec2, req2)

	resp2 := rec2.Result()
	require.Equal(t, http.StatusInternalServerError, resp2.StatusCode, "wrong status code")

	err = resp2.Body.Close()
	assert.NoError(t, err)

	sink1, ok := log.SinkFromContext(sink1Ctx)
	assert.True(t, ok, "failed getting sink1")

	var sink1LogLines []json.RawMessage

	assert.NoError(t, json.Unmarshal(sink1.ToJSON(), &sink1LogLines), "failed extracting logs from sink1")

	assert.Len(t, sink1LogLines, 2, "more log lines than expected")
	assert.Contains(t, string(sink1LogLines[0]), "ONLY FOR SINK1", "missing log line")

	sink2, ok := log.SinkFromContext(sink2Ctx)
	assert.True(t, ok, "failed getting sink2")

	var sink2LogLines []json.RawMessage

	assert.NoError(t, json.Unmarshal(sink2.ToJSON(), &sink2LogLines), "failed extracting logs from sink2")

	assert.NotContains(t, string(sink2LogLines[0]), "ONLY FOR SINK1", "unexpected log line found")

	assert.Contains(t, string(sink2LogLines[0]), "https://www.pace.car/de", "missing log line")
	assert.Contains(t, string(sink2LogLines[1]), "https://www.pace.car/de", "missing log line")
	assert.Contains(t, string(sink2LogLines[2]), "http://localhost/fail", "missing log line")

	assert.Contains(t, string(sink2LogLines[3]), "sentry:category", "missing log line")
	assert.Contains(t, string(sink2LogLines[3]), "redis", "missing log line")
	assert.Contains(t, string(sink2LogLines[3]), "sentry:type", "missing log line")
	assert.Contains(t, string(sink2LogLines[3]), "error", "missing log line")
	assert.Contains(t, string(sink2LogLines[3]), "ONLY FOR SINK2", "missing log line")

	assert.Contains(t, string(sink2LogLines[4]), "Sink2 Test Error, IGNORE", "missing log line")
}

func TestHandle(t *testing.T) {
	tests := []struct {
		name         string
		ctx          context.Context
		err          error
		handlerName  string
		expectLogMsg string
	}{
		{
			name:         "handle panic error",
			ctx:          context.Background(),
			err:          NewPanicError("test panic"),
			handlerName:  "testHandler",
			expectLogMsg: "Panic",
		},
		{
			name:         "handle regular error",
			ctx:          context.Background(),
			err:          errors.New("test error"),
			handlerName:  "testHandler",
			expectLogMsg: "Error",
		},
		{
			name:         "handle error without handler name",
			ctx:          context.Background(),
			err:          errors.New("test error"),
			handlerName:  "",
			expectLogMsg: "Error",
		},
	}

	type logEntry struct {
		Level   string `json:"level"`
		Message string `json:"message"`
		Error   string `json:"error"`
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				buf   bytes.Buffer
				entry logEntry
			)

			l := log.Logger().Output(&buf)

			handle(l.WithContext(tt.ctx), tt.err, tt.handlerName)

			line, err := buf.ReadString('\n')
			require.NoError(t, err, "failed reading log line")

			err = json.Unmarshal([]byte(line), &entry)
			require.NoError(t, err, "failed unmarshalling log line")

			assert.Equal(t, logEntry{Level: "error", Message: tt.expectLogMsg, Error: tt.err.Error()}, entry, "wrong log entry")
		})
	}
}
