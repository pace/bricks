// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/12/05 by Charlotte Pröller

package servicehealthcheck

import (
	"net/http"
)

type healthHandler struct{}

func (h *healthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, res := range check(&requiredChecks) {
		if res.State == Err {
			writeResult(w, http.StatusServiceUnavailable, string(Err))
			return
		}
	}
	writeResult(w, http.StatusOK, string(Ok))
}
