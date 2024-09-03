// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.

package transport

import (
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
)

// TracingRoundTripper implements a chainable round tripper for tracing
type TracingRoundTripper struct {
	transport http.RoundTripper
}

// Transport returns the RoundTripper to make HTTP requests
func (l *TracingRoundTripper) Transport() http.RoundTripper {
	return l.transport
}

// SetTransport sets the RoundTripper to make HTTP requests
func (l *TracingRoundTripper) SetTransport(rt http.RoundTripper) {
	l.transport = rt
}

// RoundTrip executes a HTTP request with distributed tracing
func (l *TracingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		// Check the concurrency guide for more details: https://docs.sentry.io/platforms/go/concurrency/
		hub = sentry.CurrentHub().Clone()
		ctx = sentry.SetHubOnContext(ctx, hub)
	}

	options := []sentry.SpanOption{
		// Set the OP based on values from https://develop.sentry.dev/sdk/performance/span-operations/
		sentry.WithOpName("http.client"),
		sentry.ContinueFromRequest(req),
		sentry.WithTransactionSource(sentry.SourceURL),
	}

	span := sentry.StartTransaction(ctx, fmt.Sprintf("%s %s", req.Method, req.URL.Path), options...)
	defer span.Finish()

	ctx = span.Context()
	req = req.WithContext(ctx)

	resp, err := l.Transport().RoundTrip(req)

	attempt := attemptFromCtx(ctx)
	if attempt > 0 {
		span.SetData("attempt", int(attempt))
	}
	if err != nil {
		span.SetData("error", err)
		return nil, err
	}

	span.SetData("code", resp.StatusCode)

	return resp, nil
}
