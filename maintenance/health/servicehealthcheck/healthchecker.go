// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/18 Charlotte Pröller

package servicehealthcheck

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/log"
)

type HealthCheck interface {
	InitHealthCheck() error
	HealthCheck() (bool, error)
	CleanUp() error
}

// ConnectionState can be used for Health Checks
// It offers a Mutex and the Date of the last check for caching the result
type ConnectionState struct {
	LastCheck time.Time
	isHealthy bool
	err       error
	m         sync.Mutex
}

type handler struct{}

// checks contains all registered Health Checks - key:Name
var checks sync.Map

// initErrors map with all errors that happened in the initialisation of the health checks - key:Name
var initErrors sync.Map

// router is needed to dynamically add health checks
var router *mux.Router

func (cs *ConnectionState) setConnectionState(healthy bool, err error, mom time.Time) {
	cs.m.Lock()
	defer cs.m.Unlock()
	cs.isHealthy = healthy
	cs.err = err
	cs.LastCheck = mom
}

func (cs *ConnectionState) SetErrorState(err error, mom time.Time) {
	cs.setConnectionState(err == nil, err, mom)
}

func (cs *ConnectionState) SetHealthy(mom time.Time) {
	cs.setConnectionState(true, nil, mom)
}

func (cs *ConnectionState) GetState() (bool, error) {
	return cs.isHealthy, cs.err
}

func NewConnectionState() *ConnectionState {
	return &ConnectionState{}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route := r.URL.Path
	splitRoute := strings.Split(route, "/health/")

	//Check the route first
	if len(splitRoute) < 2 || splitRoute[0] != "" {
		h.writeError(w, errors.New("Route not valid"), http.StatusBadRequest, route)
		return
	}

	name := splitRoute[1]

	hcInterface, isIn := checks.Load(name)
	if !isIn {
		h.writeError(w, errors.New("Health Check not registered\n"), http.StatusNotFound, name)
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
		// to increase performance of the request set
		// content type and write status code explicitly
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n")) // nolint: gosec,errcheck
	} else {
		h.writeError(w, err, http.StatusServiceUnavailable, name)
	}
}

func (h *handler) writeError(w http.ResponseWriter, err error, errorCode int, name string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(errorCode)
	w.Write([]byte("ERR")) // nolint: gosec,errcheck
	log.Warnf("helthcheck %v was not healthy %v", name, err)
}

// RegisterHealthCheck register a healthCheck that need to be routed
// names must be uniq
func RegisterHealthCheck(hc HealthCheck, name string) {
	if _, ok := checks.Load(name); ok {
		log.Debugf("Health checks can only be added once, tried to add another health check with name %v", name)
		return
	}
	if router == nil {
		log.Errorf("Tried to add HealthCheck ( %T ) without a router", hc)
		return
	}
	checks.Store(name, hc)
	if err := hc.InitHealthCheck(); err != nil {
		log.Warnf("Error initialising health check  %T: %v", hc, err)
		initErrors.Store(name, err)
	}
	router.Handle("/health/"+name, Handler())
}

// InitialiseHealthChecker must be called so the health checker can register new health checks as routes
func InitialiseHealthChecker(r *mux.Router) {
	router = r
}

// Handler returns the health api endpoint
func Handler() http.Handler {
	return &handler{}
}
