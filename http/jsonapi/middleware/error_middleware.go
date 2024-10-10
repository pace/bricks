package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/pace/bricks/http/jsonapi/runtime"
	"github.com/pace/bricks/maintenance/log"
)

type errorMiddleware struct {
	http.ResponseWriter
	req        *http.Request
	statusCode int
	hasErr     bool
	hasBytes   bool
}

func (e *errorMiddleware) Write(b []byte) (int, error) {
	if e.hasErr {
		log.Req(e.req).Warn().Msgf("Error already sent, ignoring: %q", string(b))
		return 0, nil
	}

	repliesJSONAPI := e.Header().Get("Content-Type") == runtime.JSONAPIContentType
	requestsJSONAPI := e.req.Header.Get("Accept") == runtime.JSONAPIContentType

	if e.statusCode >= 400 && requestsJSONAPI && !repliesJSONAPI {
		if e.hasBytes {
			log.Req(e.req).Warn().Msgf("Body already contains data from previous writes: ignoring: %q", string(b))
			return 0, nil
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

func (e *errorMiddleware) WriteHeader(code int) {
	e.statusCode = code
	e.ResponseWriter.WriteHeader(code)
}

// Error is a middleware that wraps http.ResponseWriter
// such that it forces responses with status codes 4xx/5xx to have
// Content-Type: application/vnd.api+json.
func Error(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&errorMiddleware{ResponseWriter: w, req: r}, r)
	})
}
