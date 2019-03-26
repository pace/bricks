// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/03/11 by Florian Hübsch

package transport

import "net/http"

// NewDefaultTransport returns a transport with retry, jaeger and logging support.
// Has to be finalized with a HTTP round tripper `final`. If `final` is nil `http.DefaultTransport` is used as finalizer.
func NewDefaultTransport(final http.RoundTripper) *RoundTripperChain {
	c := Chain(NewRetryRoundTripper(nil), &JaegerRoundTripper{}, &LoggingRoundTripper{})
	if final == nil {
		return c.Final(http.DefaultTransport)
	}

	return c.Final(final)
}
