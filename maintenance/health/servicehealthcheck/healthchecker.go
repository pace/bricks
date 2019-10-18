// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/18 Charlotte Pröller

package servicehealthcheck

import (
	"github.com/gorilla/mux"
	"github.com/pace/bricks/maintenance/log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type HealthCheck interface {
	Name() string
	InitHealthCheck() error
	HealthCheck(currTime time.Time) (bool, error)
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
	cs.setConnectionState(false, err, mom)
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
		hc := checks[splitRoute[1]]
		if splitRoute[1] == hc.Name() {
			//If it was not possible to initialise this health check, then show the initialisation error
			if err := initErrors[hc.Name()]; err != nil {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("Not OK:\n"[:])) // nolint: gosec,errcheck
				w.Write([]byte(err.Error()[:])) // nolint: gosec,errcheck
			} else if healthy, err := hc.HealthCheck(time.Now()); healthy {
				// to increase performance of the request set
				// content type and write status code explicitly
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK\n"[:])) // nolint: gosec,errcheck
			} else {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("Not OK:\n"[:])) // nolint: gosec,errcheck
				w.Write([]byte(err.Error()[:])) // nolint: gosec,errcheck
			}
		}
	} else {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Health Check not registered\n"[:])) // nolint: gosec,errcheck
	}
}

// RegisterHealthCheck register a healthCheck that need to be routed
// names must be uniq
func RegisterHealthCheck(hc HealthCheck) {
	if checks[hc.Name()] != nil {
		log.Warnf("Health checks can only be added once, tried to add another health check with name %v", hc.Name())
		return
	}
	if router == nil {
		log.Warnf("Tried to add HealthCheck ( %T ) without a router", hc)
		return
	}
	checks[hc.Name()] = hc
	if err := hc.InitHealthCheck(); err != nil {
		log.Warnf("Error initialising HealthCheck  %T: %v", hc, err)
		initErrors[hc.Name()] = err
	}
	router.Handle("/health/"+hc.Name(), Handler())
}

// InitialiseHealthChecker must be called so the healthchecker can register new health checks as routes
func InitialiseHealthChecker(r *mux.Router) {
	router = r
}

// Handler returns the health api endpoint
func Handler() http.Handler {
	return &handler{}
}
