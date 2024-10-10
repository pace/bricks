// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package tracing

import (
	"net/http"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/zenazn/goji/web/mutil"

	_ "github.com/pace/bricks/internal/sentry"
	"github.com/pace/bricks/maintenance/util"
)

type traceHandler struct {
	next http.Handler
}

// Trace the service function handler execution.
func (h *traceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	hub := sentry.CurrentHub()

	options := []sentry.SpanOption{
		sentry.ContinueTrace(hub, r.Header.Get(sentry.SentryTraceHeader), r.Header.Get(sentry.SentryBaggageHeader)),
		sentry.WithOpName("http.server"),
		sentry.WithTransactionSource(sentry.SourceURL),
		sentry.WithSpanOrigin(sentry.SpanOriginStdLib),
	}

	transaction := sentry.StartTransaction(ctx,
		getHTTPSpanName(r),
		options...,
	)
	transaction.SetData("http.request.method", r.Method)

	ww := mutil.WrapWriter(w)

	defer func() {
		status := ww.Status()
		bytesWritten := ww.BytesWritten()
		transaction.Status = sentry.HTTPtoSpanStatus(status)

		transaction.SetData("http.response.status_code", status)
		transaction.SetData("http.response.content_length", bytesWritten)
		transaction.SetData("http.request.url", r.URL.String())

		transaction.Finish()
	}()

	hub.Scope().SetRequest(r)

	r = r.WithContext(transaction.Context())

	h.next.ServeHTTP(ww, r)
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

// getHTTPSpanName grab needed fields from *http.Request to generate a span name for `http.server` span op.
func getHTTPSpanName(r *http.Request) string {
	if r.Pattern != "" {
		// If value does not start with HTTP methods, add them.
		// The method and the path should be separated by a space.
		if parts := strings.SplitN(r.Pattern, " ", 2); len(parts) == 1 {
			return r.Method + " " + r.Pattern
		}

		return r.Pattern
	}

	return r.Method + " " + r.URL.Path
}
