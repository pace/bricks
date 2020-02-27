// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/03/11 by Florian Hübsch

package transport

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pace/bricks/maintenance/log"
)

// LoggingRoundTripper implements a chainable round tripper for logging
type LoggingRoundTripper struct {
	transport http.RoundTripper
}

// Transport returns the RoundTripper to make HTTP requests
func (l *LoggingRoundTripper) Transport() http.RoundTripper {
	return l.transport
}

// SetTransport sets the RoundTripper to make HTTP requests
func (l *LoggingRoundTripper) SetTransport(rt http.RoundTripper) {
	l.transport = rt
}

// RoundTrip executes a HTTP request with logging
func (l *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	startTime := time.Now()
	le := log.Ctx(ctx).Debug().
		Str("url", req.URL.String()).
		Str("method", req.Method).
		Str("sentry:type", "http").
		Str("sentry:category", "http")

	resp, err := l.Transport().RoundTrip(req)

	dur := float64(time.Since(startTime)) / float64(time.Millisecond)
	le = le.Float64("duration", dur)
	attempt := attemptFromCtx(ctx)
	if attempt > 0 {
		le = le.Int("attempt", int(attempt))
	}

	if err != nil {
		le.Err(err).Msg(logEventMsg(req))
		return nil, err
	}

	le.Int("status_code", resp.StatusCode).Msg(logEventMsg(req))

	return resp, nil
}

func logEventMsg(r *http.Request) string {
	return fmt.Sprintf("%s %s %s", strings.ToUpper(r.URL.Scheme), r.Method, r.URL.Host)
}
