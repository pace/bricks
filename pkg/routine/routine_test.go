// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/02/20 by Marius Neugebauer

package routine_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/pace/bricks/http/security"
	"github.com/pace/bricks/maintenance/log"
	. "github.com/pace/bricks/pkg/routine"
	"github.com/stretchr/testify/require"
)

func TestRun_catchesPanics(t *testing.T) {
	require.NotPanics(t, func() {
		waitForRun(context.Background(), func(context.Context) {
			panic("test")
		})
	})
}

func TestRun_createsNewContext(t *testing.T) {
	ctx := context.Background()
	require.NotEqual(t, ctx, contextAfterRun(ctx, nil))
}

func TestRun_transfersLogger(t *testing.T) {
	buf := bytes.Buffer{}
	logger := log.Output(&buf)
	ctx := logger.WithContext(context.Background())
	waitForRun(ctx, func(ctx context.Context) {
		log.Ctx(ctx).Debug().Msg("foobar")
	})
	require.Contains(t, buf.String(), "foobar")
}

func TestRun_transfersSink(t *testing.T) {
	var sink log.Sink
	logger := log.Logger()
	ctx := log.ContextWithSink(logger.WithContext(context.Background()), &sink)
	waitForRun(ctx, func(ctx context.Context) {
		log.Ctx(ctx).Debug().Msg("foobar")
	})
	require.Contains(t, string(sink.ToJSON()), "foobar")
}

func TestRun_transfersOAuth2Token(t *testing.T) {
	ctx := security.ContextWithToken(context.Background(), token("test-token"))
	routineCtx := contextAfterRun(ctx, nil)
	token, ok := security.GetTokenFromContext(routineCtx)
	require.True(t, ok)
	require.Equal(t, "test-token", token.GetValue())
}

func TestRun_cancelsContextAfterRoutineIsFinished(t *testing.T) {
	routineCtx := contextAfterRun(context.Background(), nil)
	require.Eventually(t, func() bool {
		return routineCtx.Err() == context.Canceled
	}, time.Second, time.Millisecond)
}

func TestRun_blocksAfterShutdown(t *testing.T) {
	exitAfterTest(t, "TestRun_blocksAfterShutdown", testRunBlocksAfterShutdown)
}

func testRunBlocksAfterShutdown(t *testing.T) {
	var endOfTest sync.WaitGroup
	endOfTest.Add(1)

	// start routine that gets canceled by the shutdown
	routineCtx := make(chan context.Context)
	Run(context.Background(), func(ctx context.Context) {
		routineCtx <- ctx
		endOfTest.Wait()
	})

	// kill this process
	err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	require.NoError(t, err)

	// wait until the routine gets canceled because of the shutdown
	<-(<-routineCtx).Done()

	// start routine
	done := make(chan struct{})
	go func() {
		waitForRun(context.Background(), nil)
		done <- struct{}{}
	}()
	select {
	case <-done:
		t.Fatal("routine started after shutdown")
	case <-time.After(10 * time.Millisecond):
		// good
	}
}

func TestRun_cancelsContextsOnSIGINT(t *testing.T) {
	t.Skip("test not working properly in docker, skipping")

	exitAfterTest(t, "TestRun_cancelsContextsOnSIGINT", func(t *testing.T) {
		testRunCancelsContextsOn(t, syscall.SIGINT)
	})
}

func TestRun_cancelsContextsOnSIGTERM(t *testing.T) {
	t.Skip("test not working properly in docker, skipping")

	exitAfterTest(t, "TestRun_cancelsContextsOnSIGTERM", func(t *testing.T) {
		testRunCancelsContextsOn(t, syscall.SIGTERM)
	})
}

func testRunCancelsContextsOn(t *testing.T, signum syscall.Signal) {
	var endOfTest, routinesStarted sync.WaitGroup
	endOfTest.Add(1)

	// start a few routines
	routineContexts := [3]context.Context{}
	routinesStarted.Add(len(routineContexts))
	for i := range routineContexts {
		i := i
		Run(context.Background(), func(ctx context.Context) {
			routineContexts[i] = ctx
			routinesStarted.Done()
			endOfTest.Wait()
		})
	}
	routinesStarted.Wait()

	// kill this process
	err := syscall.Kill(syscall.Getpid(), signum)
	require.NoError(t, err)

	// check that all contexts are canceled
	for _, ctx := range routineContexts {
		require.Eventually(t, func() bool {
			return ctx.Err() == context.Canceled
		}, time.Second, time.Millisecond)
	}

	endOfTest.Done()
}

// Run a test, but exit afterwards. This is necessary for shutdown scenarios
// after which the routine package blocks starting new routines.
func exitAfterTest(t *testing.T, name string, testFunc func(*testing.T)) {
	if os.Getenv("ROUTINE_EXIT_AFTER_TEST") == "1" {
		testFunc(t)
		os.Exit(0)
	}
	cmd := exec.Command(os.Args[0], "-test.run="+name)
	cmd.Env = append(os.Environ(), "ROUTINE_EXIT_AFTER_TEST=1")
	require.NoError(t, cmd.Run())
}

// Calls Run and returns once the routine is finished.
func waitForRun(ctx context.Context, routine func(context.Context)) {
	done := make(chan struct{})
	Run(ctx, func(ctx context.Context) {
		defer func() { done <- struct{}{} }()
		routine(ctx)
	})
	<-done
}

// Calls Run and returns the context that the routine was called with once the
// routine is finished.
func contextAfterRun(ctx context.Context, routine func(context.Context)) context.Context {
	var routineCtx context.Context
	waitForRun(ctx, func(ctx context.Context) {
		if routine != nil {
			routine(ctx)
		}
		routineCtx = ctx
	})
	return routineCtx
}

type token string

func (t token) GetValue() string {
	return string(t)
}
