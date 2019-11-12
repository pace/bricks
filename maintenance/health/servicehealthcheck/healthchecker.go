// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/18 Charlotte Pröller

package servicehealthcheck

import (
	"net/http"
	"strings"
	"sync"

	"github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/log"
)

// HealthCheck is a health check that is initialised once and that is performed
// periodically and/or spontaneously.
type HealthCheck interface {
	HealthCheck() (bool, error)
}

type Initialisable interface {
	Init() error
}

type handler struct{}

// checks contains all registered Health Checks - key:Name
var checks sync.Map

// initErrors map with all errors that happened in the initialisation of the health checks - key:Name
var initErrors sync.Map

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route := r.URL.Path
	splitRoute := strings.Split(route, "/health/")

	// Check the route first
	if len(splitRoute) != 2 || splitRoute[0] != "" {
		h.writeError(w, errors.New("route not valid"), http.StatusBadRequest, route)
		return
	}

	name := splitRoute[1]
	hcInterface, isIn := checks.Load(name)
	if !isIn {
		h.writeError(w, errors.New("health check not registered"), http.StatusNotFound, name)
		return
	}
	hc := hcInterface.(HealthCheck)

	// If it was not possible to initialise this health check, then show the initialisation error
	if val, isIn := initErrors.Load(name); isIn {
		err := val.(error)

		h.writeError(w, err, http.StatusServiceUnavailable, name)
		return
	}

	// this is the actual health check
	if healthy, err := hc.HealthCheck(); healthy {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK\n")); err != nil {
			log.Warnf("could not write output: %s", err)
		}
	} else {
		h.writeError(w, err, http.StatusServiceUnavailable, name)
	}
}

func (h *handler) writeError(w http.ResponseWriter, err error, errorCode int, name string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(errorCode)
	if _, err := w.Write([]byte("ERR")); err != nil {
		log.Warnf("could not write output: %s", err)
	}
	log.Warnf("healthcheck %q was not healthy: %v", name, err)
}

// RegisterHealthCheck registers a HealthCheck that need to be routed. The name
// must be unique. If the health check satisfies the Initialisable interface, it
// is initialised before it is added.
func RegisterHealthCheck(hc HealthCheck, name string) {
	if _, ok := checks.Load(name); ok {
		log.Debugf("tried to register health check with name %v twice", name)
		return
	}
	if initHC, ok := hc.(Initialisable); ok {
		if err := initHC.Init(); err != nil {
			log.Warnf("error initialising health check %q: %s", name, err)
			initErrors.Store(name, err)
		}
	}
	checks.Store(name, hc)
}

// Handler returns the health api endpoint
func Handler() http.Handler {
	return &handler{}
}
