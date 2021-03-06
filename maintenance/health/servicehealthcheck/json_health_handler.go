package servicehealthcheck

import (
	"encoding/json"
	"fmt"
	"net/http"

	brickserrors "github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/log"
)

type jsonHealthHandler map[string]serviceStats

type serviceStats struct {
	Status   HealthState `json:"status"`
	Required bool        `json:"required"`
	Error    string      `json:"error"`
}

func (h *jsonHealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s := log.Sink{Silent: true}
	ctx := log.ContextWithSink(r.Context(), &s)

	reqChecks := check(ctx, &requiredChecks)
	optChecks := check(ctx, &optionalChecks)

	checkResponse := make(jsonHealthHandler)

	var errors []string
	var warnings []string
	status := http.StatusOK
	for name, res := range reqChecks {
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

	for name, res := range optChecks {
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
		brickserrors.Handle(ctx, fmt.Errorf("json health handler endpoint: encoding failed: %w", err))
	}
}
