// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/05 by Vincent Landgraf

package http

import (
	"net/http/pprof"

	"github.com/gorilla/mux"
	"lab.jamit.de/pace/go-microservice/maintenance/errors"
	"lab.jamit.de/pace/go-microservice/maintenance/health"
	"lab.jamit.de/pace/go-microservice/maintenance/log"
	"lab.jamit.de/pace/go-microservice/maintenance/metrics"
	"lab.jamit.de/pace/go-microservice/maintenance/tracing"
)

// Router returns the default microservice endpoints for
// health, metrics and debuging
func Router() *mux.Router {
	r := mux.NewRouter()

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
