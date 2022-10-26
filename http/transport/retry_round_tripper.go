// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/03/15 by Florian Hübsch

package transport

import (
	"io"
	"net"
	"net/http"
	"time"

	"github.com/PuerkitoBio/rehttp"
)

const maxRetries = 9

// RetryRoundTripper implements a chainable round tripper for retrying requests
type RetryRoundTripper struct {
	retryTransport *rehttp.Transport
	transport      http.RoundTripper
}

// RetryNetErr retries errors returned by the 'net' package.
func RetryNetErr() rehttp.RetryFn {
	return func(attempt rehttp.Attempt) bool {
		if _, isNetError := attempt.Error.(*net.OpError); isNetError {
			return true
		}
		return false
	}
}

// RetryEOFErr retries only when the error is EOF
func RetryEOFErr() rehttp.RetryFn {
	return func(attempt rehttp.Attempt) bool {
		return attempt.Error == io.EOF
	}
}

// NewDefaultRetryTransport returns a new default retry transport.
func NewDefaultRetryTransport() *rehttp.Transport {
	return rehttp.NewTransport(
		nil,
		rehttp.RetryAll(
			rehttp.RetryMaxRetries(maxRetries),
			rehttp.RetryAny(
				rehttp.RetryStatuses(408, 502, 503, 504),
				RetryEOFErr(),
				RetryNetErr(),
				rehttp.RetryTemporaryErr(),
			),
		),
		rehttp.ConstDelay(100*time.Millisecond),
	)
}

// NewRetryRoundTripper returns a retry round tripper with the specified retry transport.
func NewRetryRoundTripper(rt *rehttp.Transport) *RetryRoundTripper {
	return &RetryRoundTripper{retryTransport: rt}
}

// NewDefaultRetryRoundTripper returns a retry round tripper with a
// NewDefaultRetryTransport() as transport.
func NewDefaultRetryRoundTripper() *RetryRoundTripper {
	return &RetryRoundTripper{retryTransport: NewDefaultRetryTransport()}
}

type retryWrappedTransport struct {
	transport http.RoundTripper
	attempts  int
}

func (rt *retryWrappedTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	rt.attempts++
	return rt.transport.RoundTrip(r)
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
	retryTransport := *l.retryTransport
	wrappedTransport := &retryWrappedTransport{
		transport: transportWithAttempt(l.Transport()),
	}
	retryTransport.RoundTripper = wrappedTransport
	resp, err := retryTransport.RoundTrip(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 && wrappedTransport.attempts > maxRetries {
		return nil, ErrRetryFailed
	}

	return resp, nil
}
