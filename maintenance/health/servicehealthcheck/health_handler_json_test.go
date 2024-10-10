package servicehealthcheck

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJSONHealthHandler(t *testing.T) {
	handler := JSONHealthHandler()

	warnHC := &mockHealthCheck{name: "WithWarning", healthCheckWarn: true}
	errHC := &mockHealthCheck{name: "WithErr", healthCheckErr: true}
	initErrHC := &mockHealthCheck{name: "WithInitErr", initErr: true}

	testcases := []struct {
		title      string
		requiredHC []*mockHealthCheck
		optionalHC []*mockHealthCheck
		expCode    int
		expRes     map[string]serviceStats
	}{
		{
			title:      "all required",
			requiredHC: []*mockHealthCheck{warnHC, errHC, initErrHC},
			optionalHC: []*mockHealthCheck{},
			expCode:    http.StatusServiceUnavailable,
			expRes: map[string]serviceStats{
				errHC.name: {
					Status:   Err,
					Required: true,
					Error:    "healthCheckErr",
				},
				initErrHC.name: {
					Status:   Err,
					Required: true,
					Error:    "initError",
				},
				warnHC.name: {
					Status:   Ok,
					Required: true,
					Error:    "",
				},
			},
		},
		{
			title:      "some required, some optional",
			requiredHC: []*mockHealthCheck{warnHC, errHC},
			optionalHC: []*mockHealthCheck{initErrHC},
			expCode:    http.StatusServiceUnavailable,
			expRes: map[string]serviceStats{
				errHC.name: {
					Status:   Err,
					Required: true,
					Error:    "healthCheckErr",
				},
				initErrHC.name: {
					Status:   Err,
					Required: false,
					Error:    "initError",
				},
				warnHC.name: {
					Status:   Ok,
					Required: true,
					Error:    "",
				},
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.title, func(t *testing.T) {
			resetHealthChecks()

			for _, hc := range tc.requiredHC {
				RegisterHealthCheck(hc.name, hc)
			}

			for _, hc := range tc.optionalHC {
				RegisterOptionalHealthCheck(hc, hc.name)
			}

			testRequest(t, handler, tc.expCode, func(t *testing.T, resBody []byte) {
				var res map[string]serviceStats

				err := json.Unmarshal(resBody, &res)
				require.NoError(t, err)

				for k, v := range tc.expRes {
					require.Equal(t, v.Status, res[k].Status)
					require.Equal(t, v.Required, res[k].Required)
					require.Equal(t, v.Error, res[k].Error)
				}
			})
		})
	}
}
