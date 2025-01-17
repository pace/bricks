// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package livetest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/pace/bricks/maintenance/metric"
)

func TestIntegrationExample(t *testing.T) {
	if testing.Short() {
		t.Skip()
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	err := Test(ctx, []TestFunc{
		func(t *T) {
			t.Logf("Executed test no %d", 1)
		},
		func(t *T) {
			t.Log("Executed test no 2")
		},
		func(t *T) {
			t.Fatal("Fail test no 3")
		},
		func(t *T) {
			t.Fatalf("Fail test no %d", 4)
		},
		func(t *T) {
			t.Skip("Skipping test no 5")
		},
		func(t *T) {
			t.Skipf("Skipping test no %d", 5)
		},
		func(t *T) {
			t.SkipNow()
		},
		func(t *T) {
			t.Fail()
		},
		func(t *T) {
			t.FailNow()
		},
		func(t *T) {
			t.Error("Some")
		},
		func(t *T) {
			t.Errorf("formatted")
		},
	})

	require.ErrorIs(t, err, context.DeadlineExceeded)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	resp := httptest.NewRecorder()
	metric.Handler().ServeHTTP(resp, req)

	body := resp.Body.String()

	sn := cfg.ServiceName
	if !strings.Contains(body, `pace_livetest_total{result="failed",service="`+sn+`"} 6`) ||
		!strings.Contains(body, `pace_livetest_total{result="skipped",service="`+sn+`"} 3`) ||
		!strings.Contains(body, `pace_livetest_total{result="succeeded",service="`+sn+`"} 2`) {
		t.Errorf("expected other pace_livetest_total counts, got: %v", body)
	}
}
