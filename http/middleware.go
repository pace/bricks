package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/pace/bricks/http/jsonapi/runtime"
)

type jsonApiErrorWriter struct {
	http.ResponseWriter
	req        *http.Request
	statusCode int
}

func (e *jsonApiErrorWriter) Write(b []byte) (int, error) {
	repliesJsonApi := e.Header().Get("Content-Type") == runtime.JSONAPIContentType
	requestsJsonApi := e.req.Header.Get("Accept") == runtime.JSONAPIContentType
	if e.statusCode >= 400 && requestsJsonApi && !repliesJsonApi {
		runtime.WriteError(e, e.statusCode, errors.New(strings.Trim(string(b), "\n")))
		return 0, nil
	}

	return e.ResponseWriter.Write(b)
}

func (e *jsonApiErrorWriter) WriteHeader(code int) {
	e.statusCode = code
	e.ResponseWriter.WriteHeader(code)
}

// JsonApiErrorWriterMiddleware is a middleware that wraps http.ResponseWriter
// such that it forces responses with status codes 4xx/5xx to have
// Content-Type: application/vnd.api+json
func JsonApiErrorWriterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&jsonApiErrorWriter{w, r, 0}, r)
	})
}
