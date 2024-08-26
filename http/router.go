// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package http

import (
	"net/http"
	"net/http/pprof"

	"github.com/pace/bricks/maintenance/tracing"

	"github.com/gorilla/mux"
	"github.com/pace/bricks/http/middleware"
	"github.com/pace/bricks/locale"
	"github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/health"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
	"github.com/pace/bricks/maintenance/log"
	"github.com/pace/bricks/maintenance/metric"
	redactMdw "github.com/pace/bricks/pkg/redact/middleware"
)

// Router returns the default microservice endpoints for
// health, metrics and debugging
func Router() *mux.Router {
	r := mux.NewRouter()

	r.Use(middleware.Metrics)

	// this tracing handler must be registered before the
	// logging middleware in order to have a span context
	// initialized so that we can further log tracing data
	r.Use(tracing.Handler(
		// no tracing for these prefixes
		"/metrics",
		"/health",
		"/debug",
	))

	// the logging middleware needs to be registered before the
	// error middleware to make it possible to send panics to
	// sentry. "/health" and "/metrics" are only logged to the
	// log.Sink but not to log.output (silent)
	r.Use(log.Handler("/health", "/metrics"))

	// last resort error handler
	r.Use(errors.Handler())

	// this second handler is needed in order to have access to data managed by the log handler from above
	r.Use(tracing.TraceLogHandler(
		// no tracing for these prefixes
		"/metrics",
		"/health",
		"/debug",
	))

	r.Use(locale.Handler())

	// report use of external dependencies
	r.Use(middleware.ExternalDependency)

	// report Client ID back to caller
	r.Use(middleware.ClientID)

	// support redacting of data accross the full request scope
	r.Use(redactMdw.Redact)

	// makes some infos about the request accessable from the context
	r.Use(middleware.RequestInContext)

	// for prometheus
	r.Handle("/metrics", metric.Handler())

	// health checks
	r.Handle("/health/liveness", health.HandlerLiveness())
	r.Handle("/health/readiness", health.HandlerReadiness())

	r.Handle("/health", servicehealthcheck.HealthHandler())
	r.Handle("/health/check", servicehealthcheck.ReadableHealthHandler())
	r.Handle("/health/check.json", servicehealthcheck.JSONHealthHandler())

	// for debugging purposes (e.g. deadlock, ...)
	p := r.PathPrefix("/debug/pprof").Subrouter()
	p.HandleFunc("/cmdline", pprof.Cmdline)
	p.HandleFunc("/profile", pprof.Profile)
	p.HandleFunc("/symbol", pprof.Symbol)
	p.HandleFunc("/trace", pprof.Trace)
	p.PathPrefix("/").Handler(http.HandlerFunc(pprof.Index))

	return r
}
