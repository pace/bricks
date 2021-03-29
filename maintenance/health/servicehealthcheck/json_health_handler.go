package servicehealthcheck

import (
	"encoding/json"
	"net/http"

	"github.com/pace/bricks/maintenance/log"
)

type jsonHealthHandler struct {
	Name     string      `json: "name"`
	Status   HealthState `json: "status"`
	Required bool        `json: "required"`
	Error    string      `json: "error"`
}

func (h *jsonHealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debug("JSON HTTP handler")
	s := log.Sink{Silent: true}
	ctx := log.ContextWithSink(r.Context(), &s)

	reqChecks := check(ctx, &requiredChecks)
	optChecks := check(ctx, &optionalChecks)

	checkResponse := make([]jsonHealthHandler, len(reqChecks)+len(optChecks))
	index := 0

	status := http.StatusOK
	for name, res := range reqChecks {
		scr := jsonHealthHandler{
			Name:     name,
			Status:   res.State,
			Required: true,
			Error:    "",
		}
		if res.State == Err {
			scr.Error = res.Msg
			status = http.StatusServiceUnavailable
		}
		checkResponse[index] = scr
		index++
	}

	for name, res := range optChecks {
		scr := jsonHealthHandler{
			Name:     name,
			Status:   res.State,
			Required: false,
			Error:    "",
		}
		if res.State == Err {
			scr.Error = res.Msg
			status = http.StatusServiceUnavailable
		}
		checkResponse[index] = scr
		index++
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(checkResponse)
}
