package transport

import (
	"net/http"

	"github.com/pace/bricks/maintenance/log"
)

// TraceIDRoundTripper implements a chainable round tripper for setting the Uber-Trace-Id header
type TraceIDRoundTripper struct {
	transport  http.RoundTripper
	SourceName string
}

// Transport returns the RoundTripper to make HTTP requests
func (l *TraceIDRoundTripper) Transport() http.RoundTripper {
	return l.transport
}

// SetTransport sets the RoundTripper to make HTTP requests
func (l *TraceIDRoundTripper) SetTransport(rt http.RoundTripper) {
	l.transport = rt
}

// RoundTrip executes a single HTTP transaction via Transport()
func (l *TraceIDRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	if traceID := log.TraceIDFromContext(ctx); traceID != "" {
		req.Header.Set("Uber-Trace-Id", traceID)
	}
	return l.Transport().RoundTrip(req)
}
