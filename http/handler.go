// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/05 by Vincent Landgraf

package http

import (
	"net/http/pprof"

	"github.com/gorilla/mux"
	"lab.jamit.de/pace/go-microservice/maintenance/health"
	"lab.jamit.de/pace/go-microservice/maintenance/metrics"
)

// Router returns the default microservice endpoints for
// health, metrics and debuging
func Router() *mux.Router {
	r := mux.NewRouter()

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
