package middleware

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/pace/bricks/http/jsonapi/runtime"
)

const payload = "dummy response data"

func TestErrorMiddleware(t *testing.T) {
	for _, statusCode := range []int{200, 201, 400, 402, 500, 503} {
		for _, responseContentType := range []string{"text/plain", "text/html", runtime.JSONAPIContentType} {
			r := mux.NewRouter()
			r.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", responseContentType)
				w.WriteHeader(statusCode)
				_, _ = io.WriteString(w, payload)
			}).Methods("GET")
			r.Use(Error)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/foo", nil)
			req.Header.Set("Accept", runtime.JSONAPIContentType)

			r.ServeHTTP(rec, req)

			resp := rec.Result()
			b, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
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

				err := json.Unmarshal(b, &e)
				if err != nil {
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
		w.WriteHeader(400)
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
	}).Methods("GET")
	r.Use(Error)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/foo", nil)
	req.Header.Set("Accept", runtime.JSONAPIContentType)
	r.ServeHTTP(rec, req)
	resp := rec.Result()
	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
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
		w.WriteHeader(200)
		if _, err := io.WriteString(w, payload); err != nil {
			t.Fatal(err)
		}
		jsonWriter, ok := w.(*errorMiddleware)
		if ok && !jsonWriter.hasBytes {
			t.Fatal("expected hasBytes flag to be set")
		}
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "text/plain")
		_, _ = io.WriteString(w, payload) // will get discarded
	}).Methods("GET")
	r.Use(Error)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/foo", nil)
	req.Header.Set("Accept", runtime.JSONAPIContentType)
	r.ServeHTTP(rec, req)
	resp := rec.Result()
	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if payload != string(b) {
		t.Fatalf("bad response body, expected %q, got %q", payload, string(b))
	}
}
