// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/07 by Vincent Landgraf

package tracing

import (
	"fmt"
	"io"
	"net/http"

	opentracing "github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
	"github.com/pace/bricks/maintenance/log"
	"github.com/pace/bricks/maintenance/tracing/wire"
	"github.com/pace/bricks/maintenance/util"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics/prometheus"
	"github.com/zenazn/goji/web/mutil"
)

// Closer can be used in shutdown hooks to ensure that the internal queue of
// the Reporter is drained and all buffered spans are submitted to collectors.
var Closer io.Closer

// Tracer implementation that reports tracing to Jaeger
var Tracer opentracing.Tracer

func init() {
	cfg, err := config.FromEnv()
	if err != nil {
		log.Warnf("Unable to load Jaeger config from ENV: %v", err)
		return
	}
	if cfg.ServiceName == "" {
		log.Warn("Using Jaeger noop tracer since no JAEGER_SERVICE_NAME is present")
		return
	}

	Tracer, Closer, err = cfg.NewTracer(
		config.Metrics(prometheus.New()),
	)
	opentracing.SetGlobalTracer(Tracer)
	if err != nil {
		log.Fatal(err)
	}
}

type traceHandler struct {
	next http.Handler
}

// Trace the service function handler execution
func (h *traceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	wireContext, _ := wire.FromWire(r)
	_, ctx = opentracing.StartSpanFromContext(ctx, fmt.Sprintf("ServeHTTP Method: %s Path: %s", r.Method, r.URL.Path),
		opentracing.ChildOf(wireContext))
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
	ctx := r.Context()

	handlerSpan := opentracing.SpanFromContext(ctx)
	handlerSpan.LogFields(olog.String("req_id", log.RequestID(r)),
		olog.String("path", r.URL.Path),
		olog.String("method", r.Method))
	ww := mutil.WrapWriter(w)
	h.next.ServeHTTP(ww, r.WithContext(ctx))
	handlerSpan.LogFields(olog.Int("bytes", ww.BytesWritten()), olog.Int("status_code", ww.Status()))
	handlerSpan.Finish()
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

// Request augments an outgoing request for further tracing
func Request(r *http.Request) *http.Request {
	// check if the request contains a span
	span := opentracing.SpanFromContext(r.Context())
	if span == nil {
		return r
	}

	// inject tracing info for next request containing the span
	err := opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))
	if err != nil {
		log.Warnf("Request tracing injection failed: %v", err)
	}

	return r
}
