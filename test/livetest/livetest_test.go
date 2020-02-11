// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/02/01 by Vincent Landgraf

package livetest

import (
	"context"
	"testing"
	"time"

	"github.com/pace/bricks/test/testmetrics"
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
	if err != context.DeadlineExceeded {
		t.Error(err)
		return
	}

	sn := cfg.ServiceName
	suite := testmetrics.Setup(t, "test metrics",
		`pace_livetest_total{result="failed",service="`+sn+`"} 6`,
		`pace_livetest_total{result="skipped",service="`+sn+`"} 3`,
		`pace_livetest_total{result="succeeded",service="`+sn+`"} 2`,
	)
	suite.Run()
}
