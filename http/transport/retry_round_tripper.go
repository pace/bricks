// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/03/15 by Florian Hübsch

package transport

import (
	"errors"
	"github.com/streadway/handy/retry"
	"net/http"
	"time"

	"github.com/PuerkitoBio/rehttp"
	// "github.com/streadway/handy/retry"
)

// RetryRoundTripper implements a chainable round tripper for retrying requests
type RetryRoundTripper struct {
	retryTransport *rehttp.Transport
	transport      http.RoundTripper
}

// NewDefaultRetryTransport returns a new default retry transport.
func NewDefaultRetryTransport() *rehttp.Transport {
	// Retry: retry.All(Context(), retry.Max(9), retry.EOF(), retry.Net(), retry.Temporary(), RetryCodes(408, 502, 503, 504)),
	return rehttp.NewTransport(
		nil,
		rehttp.RetryAll(rehttp.RetryMaxRetries(9), rehttp.RetryStatuses(408, 502, 503, 504)), // max 3 retries for Temporary errors
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
