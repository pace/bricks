package servicehealthcheck

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pace/bricks/maintenance/log"
)

type serviceStats struct {
	Status   HealthState `json:"status"`
	Required bool        `json:"required"`
	Error    string      `json:"error"`
}

// JSONHealthHandler return health endpoint with all details about service health. This handler checks
// all health checks. The response body contains a JSON formatted array with every service (required or optional)
// and the detailed health checks about them.
func JSONHealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		checkResponse := make(map[string]serviceStats)

		var errors []string
		var warnings []string
		status := http.StatusOK
		for name, res := range checksResults(&requiredChecks) {
			scr := serviceStats{
				Status:   res.State,
				Required: true,
				Error:    "",
			}
			if res.State == Err {
				scr.Error = res.Msg
				status = http.StatusServiceUnavailable
				errors = append(errors, fmt.Sprintf("%s: %s", name, res.Msg))
			} else if res.State == Warn {
				warnings = append(warnings, fmt.Sprintf("%s: %s", name, res.Msg))
			}
			checkResponse[name] = scr
		}

		for name, res := range checksResults(&optionalChecks) {
			scr := serviceStats{
				Status:   res.State,
				Required: false,
				Error:    "",
			}
			if res.State == Err {
				scr.Error = res.Msg
				status = http.StatusServiceUnavailable
			}
			checkResponse[name] = scr
		}

		if len(errors) > 0 {
			log.Logger().Info().Strs("errors", errors).Strs("warnings", warnings).Msg("Health check failed")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		err := json.NewEncoder(w).Encode(checkResponse)
		if err != nil {
			log.Warnf("json health handler endpoint: encoding failed: %v", err)
		}
	}
}
