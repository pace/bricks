// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/03/16 by Florian Hübsch

package transport

import (
	"context"
	"net/http"
	"sync"
)

type ctxkey string
type attempt int

var attemptKey = ctxkey("attempt")

type attemptRoundTripper struct {
	transport http.RoundTripper
	attemptMu sync.Mutex
	attempt   attempt
}

func (l *attemptRoundTripper) Transport() http.RoundTripper {
	return l.transport
}

func (l *attemptRoundTripper) SetTransport(rt http.RoundTripper) {
	l.transport = rt
}

func (l *attemptRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	l.attemptMu.Lock()
	l.attempt++
	ctx := context.WithValue(req.Context(), attemptKey, l.attempt)
	l.attemptMu.Unlock()

	return l.Transport().RoundTrip(req.WithContext(ctx))
}

func attemptFromCtx(ctx context.Context) attempt {
	a, ok := ctx.Value(attemptKey).(attempt)
	if !ok {
		return attempt(0)
	}
	return attempt(a)
}

func transportWithAttempt(rt http.RoundTripper) http.RoundTripper {
	ar := &attemptRoundTripper{attempt: 0}
	ar.SetTransport(rt)
	return ar
}
