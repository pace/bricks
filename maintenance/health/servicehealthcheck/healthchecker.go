// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/18 Charlotte Pröller

package servicehealthcheck

import (
	"github.com/gorilla/mux"
	"github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/log"
	"net/http"
	"strings"
	"sync"
	"time"
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

// checks map with all health checks, key: Name of the check
var checks = make(map[string]HealthCheck)
var checksLock = sync.RWMutex{}

// initErrors map with all errors that happened in the initialisation of the health checks
var initErrors = make(map[string]error)
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
	return &ConnectionState{m: sync.Mutex{}}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route := r.URL.Path
	splitRoute := strings.Split(route, "/health/")
	if len(splitRoute) == 2 && splitRoute[0] == "" && checks[splitRoute[1]] != nil {
		checksLock.RLock()
		defer checksLock.RUnlock()
		name := splitRoute[1]
		if hc := checks[name]; hc != nil {
			//If it was not possible to initialise this health check, then show the initialisation error
			if err := initErrors[name]; err != nil {
				h.writeError(w, err, http.StatusServiceUnavailable, name)
			} else if healthy, err := hc.HealthCheck(); healthy {
				// to increase performance of the request set
				// content type and write status code explicitly
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK\n")) // nolint: gosec,errcheck
			} else {
				h.writeError(w, err, http.StatusServiceUnavailable, name)
			}
		} else {
			h.writeError(w, errors.New("Health Check not registered\n"), http.StatusNotFound, name)
		}
	} else {
		h.writeError(w, errors.New("Route not Valid\n"), http.StatusBadRequest, route)
	}
}

func (h *handler) writeError(w http.ResponseWriter, err error, errorCode int, name string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(errorCode)
	w.Write([]byte("Error \n")) // nolint: gosec,errcheck
	log.Warnf("helthcheck %v was not healthy %v", name, err)
}

// RegisterHealthCheck register a healthCheck that need to be routed
// names must be uniq
func RegisterHealthCheck(hc HealthCheck, name string) {
	if checks[name] != nil {
		log.Debugf("Health checks can only be added once, tried to add another health check with name %v", name)
		return
	}
	if router == nil {
		log.Warnf("Tried to add HealthCheck ( %T ) without a router", hc)
		return
	}
	checksLock.Lock()
	defer checksLock.Unlock()
	checks[name] = hc
	if err := hc.InitHealthCheck(); err != nil {
		log.Warnf("Error initialising HealthCheck  %T: %v", hc, err)
		initErrors[name] = err
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
