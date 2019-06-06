package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/pace/bricks/http/jsonapi/runtime"
	"github.com/pace/bricks/maintenance/log"
)

type jsonApiErrorWriter struct {
	http.ResponseWriter
	req        *http.Request
	statusCode int
	hasErr     bool
	hasBytes   bool
}

var errBadWriteOrder = errors.New("cannot encode jsonapi error because of previous writes")

func (e *jsonApiErrorWriter) Write(b []byte) (int, error) {
	if e.hasErr {
		log.Req(e.req).Warn().Msgf("Error already sent, ignoring: %q", string(b))
		return 0, nil
	}
	repliesJsonApi := e.Header().Get("Content-Type") == runtime.JSONAPIContentType
	requestsJsonApi := e.req.Header.Get("Accept") == runtime.JSONAPIContentType
	if e.statusCode >= 400 && requestsJsonApi && !repliesJsonApi {
		if e.hasBytes {
			return 0, errBadWriteOrder
		}
		e.hasErr = true
		runtime.WriteError(e.ResponseWriter, e.statusCode, errors.New(strings.Trim(string(b), "\n")))
		return 0, nil
	}

	n, err := e.ResponseWriter.Write(b)
	if err == nil && n > 0 {
		e.hasBytes = true
	}
	return n, err
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
		next.ServeHTTP(&jsonApiErrorWriter{ResponseWriter: w, req: r}, r)
	})
}
