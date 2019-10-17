package servicehealthcheck

import (
	"github.com/gorilla/mux"
	"github.com/pace/bricks/maintenance/log"
	"net/http"
	"time"
)

type HealthCheck interface {
	Name() string
	InitHealthCheck() error
	HealthCheck(currTime time.Time) error
	CleanUp() error
}

type handler struct{}

var checks []HealthCheck
var initErrors = make(map[string]error)
var router *mux.Router

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, hc := range checks {
		if err := initErrors[hc.Name()]; err != nil {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Not OK:\n"[:])) // nolint: gosec,errcheck
			w.Write([]byte(err.Error()[:])) // nolint: gosec,errcheck
		} else if err := hc.HealthCheck(time.Now()); err == nil {
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
		return
	}
}

// RegisterHealthCheck register a healthCheck that need to be routed
// names must be uniq
func RegisterHealthCheck(hc HealthCheck) {
	for _, val := range checks {
		if val.Name() == hc.Name() {
			log.Warnf("Health checks can only be added once, tried to add another health check with name %v", val.Name())
			return
		}
	}
	if router == nil {
		log.Warnf("Tried to add HealthCheck ( %T ) without a router", hc)
		return
	}
	checks = append(checks, hc)
	if err := hc.InitHealthCheck(); err != nil {
		log.Warnf("Error initialising HealthCheck  %T: %v", hc, err)
		initErrors[hc.Name()] = err
	}
	router.Handle("/health/"+hc.Name(), Handler())
}

func InitialiseHealthChecker(r *mux.Router) {
	router = r
}

// Handler returns the health api endpoint
func Handler() http.Handler {
	return &handler{}
}
