package servicehealthcheck

import (
	"github.com/gorilla/mux"
	"github.com/pace/bricks/maintenance/log"
	"net/http"
	"time"
)

type HealthCheck interface {
	Name() string
	InitHealthcheck() error
	HealthCheck(currTime time.Time) error
	CleanUp() error
}

type handler struct{}

var checks []HealthCheck
var router *mux.Router

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, hc := range checks {
		if err := hc.HealthCheck(time.Now()); err == nil {
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
}

// RegisterHealthCheck register a healthCheck that need to be routed
func RegisterHealthCheck(hc HealthCheck) {
	if router == nil {
		log.Warnf("Tryied to add HealthCheck ( %T ) without a router", hc)
		return
	}
	checks = append(checks, hc)
	if err := hc.InitHealthcheck(); err != nil {
		log.Warnf("Error initialising HealthCheck  %T: %v", hc, err)
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
