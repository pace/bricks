// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.

package transport

import (
	"net/http"
)

// RequestSourceRoundTripper implements a chainable round tripper for setting the Request-Source header
type RequestSourceRoundTripper struct {
	transport  http.RoundTripper
	SourceName string
}

// Transport returns the RoundTripper to make HTTP requests
func (l *RequestSourceRoundTripper) Transport() http.RoundTripper {
	return l.transport
}

// SetTransport sets the RoundTripper to make HTTP requests
func (l *RequestSourceRoundTripper) SetTransport(rt http.RoundTripper) {
	l.transport = rt
}

// RoundTrip executes a single HTTP transaction via Transport()
func (l *RequestSourceRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Request-Source", l.SourceName)
	return l.Transport().RoundTrip(req)
}
