// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.

package servicehealthcheck

import (
	"fmt"
	"net/http"

	"github.com/pace/bricks/maintenance/log"
)

// HealthHandler returns the health endpoint for transactional processing. This Handler only checks
// the required health checks and returns ERR and 503 or OK and 200.
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		var errors []string
		var warnings []string
		for name, res := range checksResults(&requiredChecks) {
			if res.State == Err {
				errors = append(errors, fmt.Sprintf("%s: %s", name, res.Msg))
			} else if res.State == Warn {
				warnings = append(warnings, fmt.Sprintf("%s: %s", name, res.Msg))
			}
		}
		if len(errors) > 0 {
			log.Logger().Info().Strs("errors", errors).Strs("warnings", warnings).Msg("Health check failed")
			msg := fmt.Sprintf("ERR: %d errors and %d warnings", len(errors), len(warnings))
			writeResult(w, http.StatusServiceUnavailable, msg)
			return
		}
		writeResult(w, http.StatusOK, string(Ok))
	}
}

func writeResult(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	if _, err := fmt.Fprint(w, body); err != nil {
		log.Warnf("could not write output: %s", err)
	}
}
