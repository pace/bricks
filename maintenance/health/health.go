// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/05 by Vincent Landgraf

// Package health implements a simple but performant handler
// that will be invoked by the loadbalancer frequently
package health

import "net/http"

type handler struct{}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// to increase performance of the request set
	// content type and write status code explicitly
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"[:]))
}

// Handler returns the health api endpoint
func Handler() http.Handler {
	return &handler{}
}
