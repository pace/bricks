// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/12/05 by Charlotte Pröller

package servicehealthcheck

import (
	"fmt"
	"net/http"
	"strconv"
)

// saves length of the longest name for the column width in the table. 20 characters width is the default
var longestCheckName = 20

type readableHealthHandler struct{}

func (h *readableHealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	reqChecks := check(&requiredChecks)
	optChecks := check(&optionalChecks)

	status := http.StatusOK
	table := "%-" + strconv.Itoa(longestCheckName) + "s   %-3s   %s\n"
	body := "Required Services: \n"
	for name, res := range reqChecks {
		body += fmt.Sprintf(table, name, res.State, res.Msg)
		if res.State == Err {
			status = http.StatusServiceUnavailable
		}
	}
	body += "Optional Services: \n"
	for name, res := range optChecks {
		body += fmt.Sprintf(table, name, res.State, res.Msg)
		if res.State == Err {
			status = http.StatusServiceUnavailable
		}
	}

	writeResult(w, status, body)
}
