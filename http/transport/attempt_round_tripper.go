// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.

package transport

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/getsentry/sentry-go"
)

type ctxkey int

var attemptKey ctxkey

type attemptRoundTripper struct {
	transport http.RoundTripper
	attempt   int32
}

func (l *attemptRoundTripper) Transport() http.RoundTripper {
	return l.transport
}

func (l *attemptRoundTripper) SetTransport(rt http.RoundTripper) {
	l.transport = rt
}

func (l *attemptRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	a := atomic.AddInt32(&l.attempt, 1)
	ctx := context.WithValue(req.Context(), attemptKey, a)

	resp, err := l.Transport().RoundTrip(req.WithContext(ctx))

	if a > 0 {
		transaction := sentry.TransactionFromContext(req.Context())
		if transaction != nil {
			transaction.SetData("attempt", fmt.Sprintf("%d", a))
		}
	}

	return resp, err
}

func attemptFromCtx(ctx context.Context) int32 {
	a, ok := ctx.Value(attemptKey).(int32)
	if !ok {
		return 0
	}
	return a
}

func transportWithAttempt(rt http.RoundTripper) http.RoundTripper {
	ar := &attemptRoundTripper{attempt: 0}
	ar.SetTransport(rt)
	return ar
}
