package http

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/pace/bricks/http/jsonapi/runtime"
	"github.com/stretchr/testify/assert"
)

func TestJsonApiErrorMiddleware(t *testing.T) {
	payload := "dummy response data"
	for _, statusCode := range []int{200, 201, 400, 402, 500, 503} {
		for _, responseContentType := range []string{"text/plain", "text/html", runtime.JSONAPIContentType} {
			r := mux.NewRouter()
			r.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", responseContentType)
				w.WriteHeader(statusCode)
				io.WriteString(w, payload)
			}).Methods("GET")
			r.Use(JsonApiErrorWriterMiddleware)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/foo", nil)
			req.Header.Set("Accept", runtime.JSONAPIContentType)

			r.ServeHTTP(rec, req)

			resp := rec.Result()
			b, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			assert.NoError(t, err)

			assert.Equal(t, statusCode, resp.StatusCode)
			if resp.StatusCode < 400 || responseContentType == runtime.JSONAPIContentType {
				assert.Equal(t, payload, string(b))
			} else {
				var e struct {
					List runtime.Errors `json:"errors"`
				}

				err := json.Unmarshal(b, &e)
				if err != nil {
					fmt.Println(string(b))
				}
				assert.NoError(t, err)
				assert.Len(t, e.List, 1)
				assert.Equal(t, payload, e.List[0].Title)
			}
		}

	}

}
