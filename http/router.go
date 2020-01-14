// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/05 by Vincent Landgraf

package http

import (
	"net/http"
	"net/http/pprof"

	"github.com/gorilla/mux"
	"github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/health"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
	"github.com/pace/bricks/maintenance/log"
	"github.com/pace/bricks/maintenance/metric"
	"github.com/pace/bricks/maintenance/tracing"
	"github.com/pace/bricks/maintenance/util"
)

// Router returns the default microservice endpoints for
// health, metrics and debugging
func Router() *mux.Router {
	r := mux.NewRouter()

	r.Use(metricsMiddleware)

	// last resort error handler
	r.Use(errors.Handler())

	// for logging
	r.Use(log.Handler(util.WithoutPrefixes("/health")))

	r.Use(tracing.Handler(
		// no tracing for these prefixes
		util.WithoutPrefixes("/metrics",
			"/health",
			"/debug",
		)))

	// for prometheus
	r.Handle("/metrics", metric.Handler())

	// health checks
	r.Handle("/health/liveness", health.HandlerLiveness())
	r.Handle("/health/readiness", health.HandlerReadiness())

	r.Handle("/health", servicehealthcheck.HealthHandler())
	r.Handle("/health/check", servicehealthcheck.ReadableHealthHandler())

	// for debugging purposes (e.g. deadlock, ...)
	p := r.PathPrefix("/debug/pprof").Subrouter()
	p.HandleFunc("/cmdline", pprof.Cmdline)
	p.HandleFunc("/profile", pprof.Profile)
	p.HandleFunc("/symbol", pprof.Symbol)
	p.HandleFunc("/trace", pprof.Trace)
	p.PathPrefix("/").Handler(http.HandlerFunc(pprof.Index))

	return r
}
