// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/03/14 by Florian Hübsch

package transport

import (
	"fmt"
	"net/http"

	"github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
)

// JaegerRoundTripper implements a chainable round tripper for tracing
type JaegerRoundTripper struct {
	transport http.RoundTripper
}

// Transport returns the RoundTripper to make HTTP requests
func (l *JaegerRoundTripper) Transport() http.RoundTripper {
	return l.transport
}

// SetTransport sets the RoundTripper to make HTTP requests
func (l *JaegerRoundTripper) SetTransport(rt http.RoundTripper) {
	l.transport = rt
}

// RoundTrip executes a HTTP request with distributed tracing
func (l *JaegerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	operationName := fmt.Sprintf("%s %s", req.Method, req.URL.Path)
	span, ctx := opentracing.StartSpanFromContext(req.Context(), operationName)
	defer span.Finish()

	resp, err := l.Transport().RoundTrip(req.WithContext(ctx))

	attempt := attemptFromCtx(ctx)
	if attempt > 0 {
		span.LogFields(olog.Int("attempt", int(attempt)))
	}
	if err != nil {
		span.LogFields(olog.Error(err))
		return nil, err
	}

	span.LogFields(olog.Int("code", resp.StatusCode))

	return resp, nil
}
