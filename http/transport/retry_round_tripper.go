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

// RetryRoundTripper implements a chainable round tripper for retrying requests
type RetryRoundTripper struct {
	retryTransport *rehttp.Transport
	transport      http.RoundTripper
}

// AbortContextDone aborts if the request's context is finished
func AbortContextDone() rehttp.RetryFn {
	return func(attempt rehttp.Attempt) bool {
		ctx := attempt.Request.Context()
		select {
		case <-ctx.Done():
			return false
		default:
			return true
		}
	}
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
		if attempt.Error == io.EOF {
			return true
		}
		return false
	}
}

// NewDefaultRetryTransport returns a new default retry transport.
func NewDefaultRetryTransport() *rehttp.Transport {
	// Retry: retry.All(Context(), retry.Max(9), retry.EOF(), retry.Net(), retry.Temporary(), RetryCodes(408, 502, 503, 504)),
	return rehttp.NewTransport(
		nil,
		rehttp.RetryAny(
			rehttp.RetryAll(
				AbortContextDone(),
				rehttp.RetryStatuses(408, 502, 503, 504),
			),
			RetryEOFErr(),
			RetryNetErr(),
			rehttp.RetryTemporaryErr(),
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
	retryTransport.RoundTripper = transportWithAttempt(l.Transport())
	resp, err := retryTransport.RoundTrip(req)

	if err != nil {
		return nil, err
	}

	return resp, nil
}
