// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package transport

import (
	"net/http"
	"time"

	"github.com/pace/bricks/http/middleware"
)

// ExternalDependencyRoundTripper greps external dependency headers and
// attach them to the currect context
type ExternalDependencyRoundTripper struct {
	name      string
	transport http.RoundTripper
}

func NewExternalDependencyRoundTripper(name string) *ExternalDependencyRoundTripper {
	return &ExternalDependencyRoundTripper{name: name}
}

// Transport returns the RoundTripper to make HTTP requests
func (l *ExternalDependencyRoundTripper) Transport() http.RoundTripper {
	return l.transport
}

// SetTransport sets the RoundTripper to make HTTP requests
func (l *ExternalDependencyRoundTripper) SetTransport(rt http.RoundTripper) {
	l.transport = rt
}

// RoundTrip executes a single HTTP transaction via Transport()
func (l *ExternalDependencyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := l.Transport().RoundTrip(req)
	elapsed := time.Since(start)

	ec := middleware.ExternalDependencyContextFromContext(req.Context())
	if ec != nil {
		if l.name != "" {
			ec.AddDependency(l.name, elapsed)
		}

		if resp != nil {
			header := resp.Header.Get(middleware.ExternalDependencyHeaderName)
			if header != "" {
				ec.Parse(header)
			}
		}
	}

	return resp, err
}
