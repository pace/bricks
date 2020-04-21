// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/03/15 by Florian Hübsch

package transport

import (
	"errors"
	"net/http"
	"time"

	"github.com/streadway/handy/retry"
)

// RetryRoundTripper implements a chainable round tripper for retrying requests
type RetryRoundTripper struct {
	retryTransport *retry.Transport
	transport      http.RoundTripper
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

// NewDefaultRetryTransport returns a new default retry transport.
func NewDefaultRetryTransport() *retry.Transport {
	return &retry.Transport{
		Delay: retry.Constant(100 * time.Millisecond),
		Retry: retry.All(Context(), retry.Max(9), retry.EOF(), retry.Net(), retry.Temporary(), RetryCodes(408, 502, 503, 504)),
	}
}

// NewRetryRoundTripper returns a retry round tripper with the specified retry transport.
func NewRetryRoundTripper(rt *retry.Transport) *RetryRoundTripper {
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
	retryTransport.Next = transportWithAttempt(l.Transport())
	resp, err := retryTransport.RoundTrip(req)

	var maxRetriesError retry.MaxError
	if err != nil {
		switch {
		case errors.As(err, &maxRetriesError):
			return nil, ErrRetryFailed
		default:
			return nil, err
		}
	}

	return resp, nil
}
