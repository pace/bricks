// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package livetest_test

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/pace/bricks/test/livetest"
)

func ExampleTest() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	err := livetest.Test(ctx, []livetest.TestFunc{
		func(t *livetest.T) {
			t.Logf("Executed test no %d", 1)
		},
		func(t *livetest.T) {
			t.Log("Executed test no 2")
		},
		func(t *livetest.T) {
			t.Fatal("Fail test no 3")
		},
		func(t *livetest.T) {
			t.Fatalf("Fail test no %d", 4)
		},
		func(t *livetest.T) {
			t.Skip("Skipping test no 5")
		},
		func(t *livetest.T) {
			t.Skipf("Skipping test no %d", 5)
		},
		func(t *livetest.T) {
			t.SkipNow()
		},
		func(t *livetest.T) {
			t.Fail()
		},
		func(t *livetest.T) {
			t.FailNow()
		},
		func(t *livetest.T) {
			t.Error("Some")
		},
		func(t *livetest.T) {
			t.Errorf("formatted")
		},
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		log.Fatal(err)
	}
	// Output:
}
