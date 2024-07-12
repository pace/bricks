// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package transport

import (
	"net/http"

	"github.com/pace/bricks/locale"
)

// LocaleRoundTripper implements a chainable round tripper for locale forwarding
type LocaleRoundTripper struct {
	transport http.RoundTripper
}

// Transport returns the RoundTripper to make HTTP requests
func (l *LocaleRoundTripper) Transport() http.RoundTripper {
	return l.transport
}

// SetTransport sets the RoundTripper to make HTTP requests
func (l *LocaleRoundTripper) SetTransport(rt http.RoundTripper) {
	l.transport = rt
}

// RoundTrip executes a HTTP request with logging
func (l *LocaleRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	loc, ok := locale.FromCtx(req.Context())
	if ok {
		return l.Transport().RoundTrip(loc.Request(req))
	} else {
		return l.Transport().RoundTrip(req)
	}
}
