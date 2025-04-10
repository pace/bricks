package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/pace/bricks/http/jsonapi/runtime"
)

const payload = "dummy response data"

func TestErrorMiddleware(t *testing.T) {
	for _, statusCode := range []int{http.StatusOK, http.StatusCreated, http.StatusBadRequest, 402, 500, 503} {
		for _, responseContentType := range []string{"text/plain", "text/html", runtime.JSONAPIContentType} {
			r := mux.NewRouter()
			r.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", responseContentType)
				w.WriteHeader(statusCode)
				_, _ = io.WriteString(w, payload)
			}).Methods(http.MethodGet)
			r.Use(Error)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/foo", nil)
			req.Header.Set("Accept", runtime.JSONAPIContentType)

			r.ServeHTTP(rec, req)

			resp := rec.Result()
			b, err := io.ReadAll(resp.Body)

			defer func() {
				err := resp.Body.Close()
				assert.NoError(t, err)
			}()

			if err != nil {
				t.Fatal(err)
			}

			if statusCode != resp.StatusCode {
				t.Fatalf("status codes differ: expected %v, got %v", statusCode, resp.StatusCode)
			}

			if resp.StatusCode < 400 || responseContentType == runtime.JSONAPIContentType {
				if payload != string(b) {
					t.Fatalf("payloads differ: expected %v, got %v", payload, string(b))
				}
			} else {
				var e struct {
					List runtime.Errors `json:"errors"`
				}

				if err := json.Unmarshal(b, &e); err != nil {
					t.Fatal(err)
				}

				if len(e.List) != 1 {
					t.Fatalf("expected only one record, got %v", len(e.List))
				}

				if payload != e.List[0].Title {
					t.Fatalf("error titles differ: expected %v, got %v", payload, e.List[0].Title)
				}
			}
		}
	}
}

func TestJsonApiErrorMiddlewareMultipleErrorWrite(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "text/html")
		if _, err := io.WriteString(w, payload); err != nil {
			t.Fatal(err)
		}
		if jsonWriter, ok := w.(*errorMiddleware); ok && !jsonWriter.hasErr {
			t.Fatal("expected hasErr flag to be set")
		}
		if _, err := io.WriteString(w, payload); err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(w, payload); err != nil {
			t.Fatal(err)
		}
	}).Methods(http.MethodGet)
	r.Use(Error)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	req.Header.Set("Accept", runtime.JSONAPIContentType)
	r.ServeHTTP(rec, req)

	resp := rec.Result()
	b, err := io.ReadAll(resp.Body)

	defer func() {
		err := resp.Body.Close()
		assert.NoError(t, err)
	}()

	if err != nil {
		t.Fatal(err)
	}

	var e struct {
		List runtime.Errors `json:"errors"`
	}

	if err := json.Unmarshal(b, &e); err != nil {
		t.Fatal(err)
	}

	if len(e.List) != 1 {
		t.Fatalf("expected only one record, got %v", len(e.List))
	}

	if payload != e.List[0].Title {
		t.Fatalf("error titles differ: expected %v, got %v", payload, e.List[0].Title)
	}
}

func TestJsonApiErrorMiddlewareInvalidWriteOrder(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := io.WriteString(w, payload); err != nil {
			t.Fatal(err)
		}
		jsonWriter, ok := w.(*errorMiddleware)
		if ok && !jsonWriter.hasBytes {
			t.Fatal("expected hasBytes flag to be set")
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "text/plain")
		_, _ = io.WriteString(w, payload) // will get discarded
	}).Methods(http.MethodGet)
	r.Use(Error)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	req.Header.Set("Accept", runtime.JSONAPIContentType)
	r.ServeHTTP(rec, req)

	resp := rec.Result()
	b, err := io.ReadAll(resp.Body)

	defer func() {
		err := resp.Body.Close()
		assert.NoError(t, err)
	}()

	if err != nil {
		t.Fatal(err)
	}

	if payload != string(b) {
		t.Fatalf("bad response body, expected %q, got %q", payload, string(b))
	}
}
