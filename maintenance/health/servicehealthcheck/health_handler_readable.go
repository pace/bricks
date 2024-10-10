// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.

package servicehealthcheck

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// saves length of the longest name for the column width in the table. 20 characters width is the default.
var longestCheckName = 20

// ReadableHealthHandler returns the health endpoint with all details about service health. This handler checks
// all health checks. The response body contains two tables (for required and optional health checks)
// with the detailed results of the health checks.
func ReadableHealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		reqChecks := checksResults(&requiredChecks)
		optChecks := checksResults(&optionalChecks)

		status := http.StatusOK
		table := "%-" + strconv.Itoa(longestCheckName) + "s   %-3s   %s\n"
		bodyBuilder := &strings.Builder{}
		bodyBuilder.WriteString("Required Services: \n")

		for name, res := range reqChecks {
			bodyBuilder.WriteString(fmt.Sprintf(table, name, res.State, res.Msg))

			if res.State == Err {
				status = http.StatusServiceUnavailable
			}
		}

		bodyBuilder.WriteString("Optional Services: \n")

		for name, res := range optChecks {
			bodyBuilder.WriteString(fmt.Sprintf(table, name, res.State, res.Msg))
			// do not change status, as this is optional
		}

		writeResult(w, status, bodyBuilder.String())
	}
}
