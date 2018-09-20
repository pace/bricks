// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/07 by Vincent Landgraf

package tracing

import (
	"net/http"

	opentracing "github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics/prometheus"
	"lab.jamit.de/pace/go-microservice/maintenance/log"
)

func init() {
	cfg, err := config.FromEnv()
	if cfg.ServiceName == "" {
		log.Warn("Using Jaeger noop tracer since no JAEGER_SERVICE_NAME is present")
		return
	}

	if err != nil {
		log.Warnf("Unable to load Jaeger config from ENV: %v", err)
		return
	}

	tracer, _, err := cfg.NewTracer(
		config.Metrics(prometheus.New()),
	)
	opentracing.SetGlobalTracer(tracer)
	if err != nil {
		log.Fatal(err)
	}
}

type tranceHandler struct {
	next http.Handler
}

// Trace the service function handler execution
func (h *tranceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var handlerSpan opentracing.Span
	wireContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("Couldn't get span from request header")
	}
	handlerSpan, ctx = opentracing.StartSpanFromContext(ctx, "ServeHTTP", opentracing.ChildOf(wireContext))
	handlerSpan.LogFields(olog.String("req_id", log.RequestID(r)))

	defer handlerSpan.Finish()

	h.next.ServeHTTP(w, r.WithContext(ctx))
}

// Handler generates a tracing handler that decodes the current trace from the wire
func Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return &tranceHandler{next: next}
	}
}
