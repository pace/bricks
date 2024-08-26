package servicehealthcheck

import (
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadableHealthHandler(t *testing.T) {
	handler := ReadableHealthHandler()

	longName := &mockHealthCheck{name: "veryveryveryveryveryveryveryveryveryveryveryverylongname"}
	warn := &mockHealthCheck{name: "WithWarning", healthCheckWarn: true}
	err := &mockHealthCheck{name: "WithErr", healthCheckErr: true}
	initErr := &mockHealthCheck{name: "WithInitErr", initErr: true}

	testcases := []struct {
		title   string
		req     []*mockHealthCheck
		opt     []*mockHealthCheck
		expCode int
		expReq  []string
		expOpt  []string
	}{
		{
			title:   "Test health check readable all required",
			req:     []*mockHealthCheck{longName, warn, err, initErr},
			opt:     []*mockHealthCheck{},
			expCode: http.StatusServiceUnavailable,
			expReq: []string{
				"WithWarning                                                OK    ",
				"WithErr                                                    ERR   healthCheckErr",
				"WithInitErr                                                ERR   initError",
				"veryveryveryveryveryveryveryveryveryveryveryverylongname   OK    "},
		},
		{
			title:   "Test health check readable all optional",
			req:     []*mockHealthCheck{},
			opt:     []*mockHealthCheck{longName, warn, err, initErr},
			expCode: http.StatusOK,
			expOpt: []string{
				"WithWarning                                                OK    ",
				"WithErr                                                    ERR   healthCheckErr",
				"WithInitErr                                                ERR   initError",
				"veryveryveryveryveryveryveryveryveryveryveryverylongname   OK    "},
		},
		{
			title:   "Test health check readable ok, all duplicated and with warning",
			req:     []*mockHealthCheck{longName, longName},
			opt:     []*mockHealthCheck{longName, warn},
			expCode: http.StatusOK,
			expReq:  []string{"veryveryveryveryveryveryveryveryveryveryveryverylongname   OK    "},
			expOpt:  []string{"WithWarning                                                OK    "},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.title, func(t *testing.T) {
			resetHealthChecks()

			for _, hc := range tc.req {
				RegisterHealthCheck(hc.name, hc)
			}
			for _, hc := range tc.opt {
				RegisterOptionalHealthCheck(hc, hc.name)
			}

			testRequest(t, handler, tc.expCode, func(t *testing.T, resBody []byte) {
				// health check results are not ordered because ranging over a map is randomized,
				// so the result is split into slices of health check results
				results := strings.Split(string(resBody), "Optional Services: \n")
				reqRes := strings.Split(strings.Split(results[0], "Required Services: \n")[1], "\n")
				optRes := strings.Split(results[1], "\n")
				testListHealthChecks(t, tc.expReq, reqRes)
				testListHealthChecks(t, tc.expOpt, optRes)
			})
		})
	}
}

func testListHealthChecks(t *testing.T, expected []string, checkResult []string) {
	t.Helper()

	// checkResult contains an empty string because of the splitting
	require.Equal(t, len(expected), len(checkResult)-1, "The amount of health check results in the response body needs to be equal to the amount of expected health check results.")

	// health check results are not ordered because ranging over a map is randomized,
	// so the rows needs to be sorted for easy comprehension
	sort.Stable(sort.StringSlice(expected))
	sort.Stable(sort.StringSlice(checkResult))

	for i := range expected {
		require.Equal(t, expected[i], checkResult[i+1], "The entry of the health check table is wrong")
	}
}
