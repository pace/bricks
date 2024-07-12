// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package transport

import (
	"net/http"

	"github.com/pace/bricks/maintenance/log"
)

// RequestIDRoundTripper implements a chainable round tripper for setting the Request-Source header
type RequestIDRoundTripper struct {
	transport  http.RoundTripper
	SourceName string
}

// Transport returns the RoundTripper to make HTTP requests
func (l *RequestIDRoundTripper) Transport() http.RoundTripper {
	return l.transport
}

// SetTransport sets the RoundTripper to make HTTP requests
func (l *RequestIDRoundTripper) SetTransport(rt http.RoundTripper) {
	l.transport = rt
}

// RoundTrip executes a single HTTP transaction via Transport()
func (l *RequestIDRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	if reqID := log.RequestIDFromContext(ctx); reqID != "" {
		req.Header.Set("Request-Id", reqID)
	}
	return l.Transport().RoundTrip(req)
}
