// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/05 by Vincent Landgraf

// Package health implements simple checks for readiness and liveness
// that will be invoked by the loadbalancer frequently
package health

import (
	"net/http"

	"github.com/pace/bricks/maintenance/log"
)

type handler struct {
	check func(http.ResponseWriter, *http.Request)
}

var readinessCheck = &handler{check: liveness}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.check(w, r)
}

// ReadinessCheck allows to set a different function for the readiness check. The default readiness check
// is the same as the liveness check and does always return OK
func SetCustomReadinessCheck(check func(http.ResponseWriter, *http.Request)) {
	readinessCheck.check = check
}

func liveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK\n"[:])); err != nil {
		log.Warnf("could not write output: %s", err)
	}
}

// HandlerLiveness returns the liveness handler that always return OK and 200
func HandlerLiveness() http.Handler {
	return &handler{check: liveness}
}

// HandlerReadiness returns the readiness handler. This handler can be configured with
// ReadinessCheck(func(http.ResponseWriter,*http.Request)), the default behavior is a liveness check
func HandlerReadiness() http.Handler {
	return readinessCheck
}
