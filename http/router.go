// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/05 by Vincent Landgraf

package http

import (
	"net/http/pprof"

	"github.com/gorilla/mux"
	"github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/health"
	"github.com/pace/bricks/maintenance/log"
	"github.com/pace/bricks/maintenance/metrics"
	"github.com/pace/bricks/maintenance/tracing"
)

// Router returns the default microservice endpoints for
// health, metrics and debugging
func Router() *mux.Router {
	r := mux.NewRouter()

	r.Use(metricsMiddleware)

	// last resort error handler
	r.Use(errors.Handler())

	// for logging
	r.Use(log.Handler())

	r.Use(tracing.Handler(
		// no tracing for these prefixes
		"/metrics",
		"/health",
		"/debug",
	))

	// for prometheus
	r.Handle("/metrics", metrics.Handler())

	// for the API gateway
	r.Handle("/health", health.Handler())

	// for debugging purposes (e.g. deadlock, ...)
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return r
}
