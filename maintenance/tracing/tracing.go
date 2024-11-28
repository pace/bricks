// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package tracing

import (
	"fmt"
	"net/http"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/pace/bricks/maintenance/log"
	"github.com/pace/bricks/maintenance/util"
	"github.com/zenazn/goji/web/mutil"
)

func init() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		Environment:      os.Getenv("ENVIRONMENT"),
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %v", err)
	}
}

type traceHandler struct {
	next http.Handler
}

// Trace the service function handler execution
func (h *traceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		// Check the concurrency guide for more details: https://docs.sentry.io/platforms/go/concurrency/
		hub = sentry.CurrentHub().Clone()
		ctx = sentry.SetHubOnContext(ctx, hub)
	}

	options := []sentry.SpanOption{
		// Set the OP based on values from https://develop.sentry.dev/sdk/performance/span-operations/
		sentry.WithOpName("http.server"),
		sentry.ContinueFromRequest(r),
		sentry.WithTransactionSource(sentry.SourceURL),
	}

	span := sentry.StartTransaction(ctx,
		fmt.Sprintf("%s %s", r.Method, r.URL.Path),
		options...,
	)

	defer span.Finish()

	ctx = span.Context()
	ww := mutil.WrapWriter(w)

	h.next.ServeHTTP(ww, r.WithContext(ctx))
}

// Handler generates a tracing handler that decodes the current trace from the wire.
// The tracing handler will not start traces for the list of ignoredPrefixes.
func Handler(ignoredPrefixes ...string) func(http.Handler) http.Handler {
	return util.NewIgnorePrefixMiddleware(func(next http.Handler) http.Handler {
		return &traceHandler{
			next: next,
		}
	}, ignoredPrefixes...)
}

type traceLogHandler struct {
	next http.Handler
}

// Trace the service function handler execution
func (h *traceLogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	span := sentry.TransactionFromContext(r.Context())

	if span != nil {
		span.SetData("req_id", log.RequestIDFromContext(r.Context()))
		span.SetData("path", r.URL.Path)
		span.SetData("method", r.Method)
	}

	ww := mutil.WrapWriter(w)

	h.next.ServeHTTP(ww, r)

	if span != nil {
		span.SetData("bytes", ww.BytesWritten())
		span.SetData("status_code", ww.Status())
	}
}

// TraceLogHandler generates a tracing handler that adds logging data to existing handler.
// The tracing handler will not start traces for the list of ignoredPrefixes.
func TraceLogHandler(ignoredPrefixes ...string) func(http.Handler) http.Handler {
	return util.NewIgnorePrefixMiddleware(func(next http.Handler) http.Handler {
		return &traceLogHandler{
			next: next,
		}
	}, ignoredPrefixes...)
}
