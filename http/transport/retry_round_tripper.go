// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/03/15 by Florian Hübsch

package transport

import (
	"net/http"
	"time"

	"github.com/streadway/handy/retry"
)

// RetryRoundTripper implements a chainable round tripper for retrying requests
type RetryRoundTripper struct {
	retryTransportFactory func() *retry.Transport
	transport             http.RoundTripper
}

// RetryCodes retries when the status code is one of the provided list
func RetryCodes(codes ...int) retry.Retryer {
	return func(a retry.Attempt) (retry.Decision, error) {
		if a.Response == nil {
			return retry.Ignore, nil
		}
		for _, code := range codes {
			if a.Response.StatusCode == code {
				return retry.Retry, nil
			}
		}
		return retry.Ignore, nil
	}
}

// Context aborts if the request's context is finished
func Context() retry.Retryer {
	return func(a retry.Attempt) (retry.Decision, error) {
		ctx := a.Request.Context()
		select {
		case <-ctx.Done():
			return retry.Abort, ctx.Err()
		default:
			return retry.Ignore, nil
		}
	}
}

// DefaultRetryTransport is used as default retry mechanism
//
// Deprecated: Use NewDefaultRetryTransport() instead. The DefaultRetryTransport
// will be removed in future versions. It is broken as its global nature means
// that the DefaultRetryTransport.Next http.RoundTripper is shared among all
// users. In previous versions of this package this would mean that the last
// user to call RetryRoundTripper.SetTransport() defines the shared
// DefaultRetryTransport.Next http.RoundTripper. This was unknowingly changed in
// v0.1.12. After that this still is at least a race condition issue as we can
// have multiple simultaneous usages, that all set DefaultRetryTransport.Next
// and indirectly call this shared state via DefaultRetryTransport.RoundTrip().
var DefaultRetryTransport = retry.Transport{
	Delay: retry.Constant(100 * time.Millisecond),
	Retry: retry.All(Context(), retry.Max(9), retry.EOF(), retry.Net(), retry.Temporary(), RetryCodes(408, 502, 503, 504)),
}

// NewDefaultRetryTransport returns a new default retry transport.
func NewDefaultRetryTransport() *retry.Transport {
	return &retry.Transport{
		Delay: retry.Constant(100 * time.Millisecond),
		Retry: retry.All(Context(), retry.Max(9), retry.EOF(), retry.Net(), retry.Temporary(), RetryCodes(408, 502, 503, 504)),
	}
}

// NewRetryRoundTripper returns a retry round tripper with the specified retry transport.
func NewRetryRoundTripper(rtf func() *retry.Transport) *RetryRoundTripper {
	return &RetryRoundTripper{retryTransportFactory: rtf}
}

// NewDefaultRetryRoundTripper returns a retry round tripper with a
// NewDefaultRetryTransport() as transport.
func NewDefaultRetryRoundTripper() *RetryRoundTripper {
	return &RetryRoundTripper{retryTransportFactory: func() *retry.Transport { return NewDefaultRetryTransport() }}
}

// Transport returns the RoundTripper to make HTTP requests
func (l *RetryRoundTripper) Transport() http.RoundTripper {
	return l.transport
}

// SetTransport sets the RoundTripper to make HTTP requests
func (l *RetryRoundTripper) SetTransport(rt http.RoundTripper) {
	l.transport = rt
}

// RoundTrip executes a HTTP request with retrying
func (l *RetryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	retryTransport := l.retryTransportFactory()
	retryTransport.Next = transportWithAttempt(l.Transport())
	resp, err := retryTransport.RoundTrip(req)

	if err != nil {
		return nil, err
	}

	return resp, nil
}
