// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package jsonapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pace/bricks/maintenance/metric"
)

func TestMetric(t *testing.T) {
	t.Run("capture metrics", func(t *testing.T) {
		t.Run("api request", func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test/1234567", nil)

			handler := func(w http.ResponseWriter, r *http.Request) {
				w = NewMetric("simple", "/test/{id}", w, r)
				w.WriteHeader(204)
			}

			handler(rec, req)
			req.Body.Close() // that's something the server does

			resp := rec.Result()
			defer resp.Body.Close()

			if resp.StatusCode != 204 {
				t.Errorf("Failed to return correct 204 response status, got: %v", resp.StatusCode)
			}
		})
		t.Run("get metrics request", func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/metrics", nil)
			metric.Handler().ServeHTTP(rec, req)

			body := rec.Body.String()
			for _, metric := range []string{
				"pace_api_http_request_total",
				"pace_api_http_request_duration_seconds",
				"pace_api_http_size_bytes",
			} {
				if !strings.Contains(body, metric) {
					t.Errorf("Expected pace api metric %q, got: %v", metric, body)
				}
			}
		})
	})

	t.Run("capture request size", func(t *testing.T) {
		t.Run("api request", func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/noop", strings.NewReader("some static request body"))

			handler := func(w http.ResponseWriter, r *http.Request) {
				NewMetric("noop", "/noop", w, r)
			}

			handler(rec, req)
			req.Body.Close() // that's something the server does
		})
		t.Run("get metrics request", func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/metrics", nil)
			metric.Handler().ServeHTTP(rec, req)

			body := rec.Body.String()
			wantMetric := `pace_api_http_size_bytes_sum{method="POST",path="/noop",service="noop",type="req"} 24`
			if !strings.Contains(body, wantMetric) {
				t.Errorf("Expected metric %q, got: %v", wantMetric, body)
			}
		})
	})

	t.Run("capture request size of stream if body is read", func(t *testing.T) {
		t.Run("api request", func(t *testing.T) {
			rec := httptest.NewRecorder()
			reqBody := strings.NewReader("some request body")
			req := httptest.NewRequest("POST", "/foobar", readerWithoutLen{reqBody})

			handler := func(w http.ResponseWriter, r *http.Request) {
				NewMetric("foobar", "/foobar", w, r)
				_, err := io.Copy(io.Discard, r.Body) // read request body
				if err != nil {
					panic(err)
				}
			}

			handler(rec, req)
			req.Body.Close() // that's something the server does
		})
		t.Run("get metrics request", func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/metrics", nil)
			metric.Handler().ServeHTTP(rec, req)

			body := rec.Body.String()
			wantMetric := `pace_api_http_size_bytes_sum{method="POST",path="/foobar",service="foobar",type="req"} 17`
			if !strings.Contains(body, wantMetric) {
				t.Errorf("Expected metric %q, got: %v", wantMetric, body)
			}
		})
	})

	t.Run("capture request size of stream if body is not read", func(t *testing.T) {
		t.Run("api request", func(t *testing.T) {
			rec := httptest.NewRecorder()
			reqBody := strings.NewReader("some request body that noone ever reads")
			req := httptest.NewRequest("POST", "/barfoo", readerWithoutLen{reqBody})

			handler := func(w http.ResponseWriter, r *http.Request) {
				NewMetric("barfoo", "/barfoo", w, r)
				// do not read request body
			}

			handler(rec, req)
			req.Body.Close() // that's something the server does
		})
		t.Run("get metrics request", func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/metrics", nil)
			metric.Handler().ServeHTTP(rec, req)

			body := rec.Body.String()
			wantMetric := `pace_api_http_size_bytes_sum{method="POST",path="/barfoo",service="barfoo",type="req"} 39`
			if !strings.Contains(body, wantMetric) {
				t.Errorf("Expected metric %q, got: %v", wantMetric, body)
			}
		})
	})

	t.Run("capture response size", func(t *testing.T) {
		t.Run("api request", func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/lalala", nil)

			handler := func(w http.ResponseWriter, r *http.Request) {
				w = NewMetric("lalala", "/lalala", w, r)
				_, err := w.Write([]byte("hehehehe"))
				if err != nil {
					panic(err)
				}
			}

			handler(rec, req)
			req.Body.Close() // that's something the server does
		})
		t.Run("get metrics request", func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/metrics", nil)
			metric.Handler().ServeHTTP(rec, req)

			body := rec.Body.String()
			wantMetric := `pace_api_http_size_bytes_sum{method="GET",path="/lalala",service="lalala",type="resp"} 8`
			if !strings.Contains(body, wantMetric) {
				t.Errorf("Expected metric %q, got: %v", wantMetric, body)
			}
		})
	})
}

// readerWithoutLen is a reader that has definitely not a Len() method
type readerWithoutLen struct {
	io.Reader
}
