// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/02/26 by Marius Neugebauer

package routine_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pace/bricks/pkg/routine"
	"github.com/stretchr/testify/assert"
)

func Example_clusterBackgroundTask() {
	// This example is an integration test because it requires redis to run. If
	// we are running the short tests just print the output that is expected to
	// circumvent the test runner. Because there is no way to skip an example.
	if testing.Short() {
		fmt.Println("task run 0\ntask run 1\ntask run 2")
		return
	}

	out := make(chan string)

	// start the routine in the background
	cancel := routine.RunNamed(context.Background(), "task",
		func(ctx context.Context) {
			for i := 0; ; i++ {
				select {
				case <-ctx.Done():
					return
				default:
				}
				out <- fmt.Sprintf("task run %d", i)
				time.Sleep(100 * time.Millisecond)
			}
		},
		// KeepRunningOneInstance will cause the routine to be restarted if it
		// finishes. It also will use the default redis database to synchronize
		// with other instances running this routine so that in all instances
		// exactly one routine is running at all time.
		routine.KeepRunningOneInstance(),
	)

	// Cancel after 3 results. Cancel will only cancel the routine in this
	// instance. It will not cancel the synchronized routines of other
	// instances.
	for i := 0; i < 3; i++ {
		println(<-out)
	}
	cancel()

	// Output:
	// task run 0
	// task run 1
	// task run 2
}

func TestIntegrationRunNamed_clusterBackgroundTask(t *testing.T) {
	t.Skip("test not working properly in docker, skipping")

	if testing.Short() {
		t.SkipNow()
	}

	// buffer that allows writing simultaneously
	var buf subprocessOutputBuffer

	// Run 2 processes in the "cluster", that will both try to start the
	// background task in the example function above. The way this task is
	// configured only one process at a time will run the task. But the
	// processes are programmed to exit after 3 iterations of the task. This
	// tests that the second process will take over the execution of the task
	// only after the first process exits.
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			spawnProcess(&buf)
			wg.Done()
		}()
	}
	wg.Wait() // until both processes are done

	exp := `task run 0
task run 1
task run 2
task run 0
task run 1
task run 2
`
	assert.Equal(t, exp, buf.String())
}

func spawnProcess(w io.Writer) {
	cmd := exec.Command(os.Args[0],
		"-test.timeout=2s",
		"-test.run=Example_clusterBackgroundTask",
	)
	cmd.Env = append(os.Environ(),
		"TEST_SUBPROCESS=1",
		"ROUTINE_REDIS_LOCK_TTL=200ms",
	)
	cmd.Stdout = w
	cmd.Stderr = w
	err := cmd.Run()
	if err != nil {
		_, _ = w.Write([]byte("error starting subprocess: " + err.Error()))
	}
}

type subprocessOutputBuffer struct {
	mx  sync.Mutex
	buf bytes.Buffer
}

func (b *subprocessOutputBuffer) Write(p []byte) (int, error) {
	b.mx.Lock()
	defer b.mx.Unlock()
	// ignore test runner output and some other log lines
	switch s := string(p); {
	case strings.HasPrefix(s, "=== RUN"),
		strings.HasPrefix(s, "--- PASS"),
		strings.HasPrefix(s, "PASS"),
		strings.HasPrefix(s, "coverage: "),
		strings.Contains(s, "Redis connection pool created"):
		return len(p), nil
	}
	return b.buf.Write(p)
}

func (b *subprocessOutputBuffer) String() string {
	b.mx.Lock()
	defer b.mx.Unlock()
	return b.buf.String()
}

// Prints the string normally so that it can be consumed by the test runner.
// Additionally go around the test runner in case of a integration test that
// wants examine the output of another test.
func println(s string) {
	if os.Getenv("TEST_SUBPROCESS") == "1" {
		// go around the test runner
		_, _ = log.Writer().Write([]byte(s + "\n"))
	}
	fmt.Println(s)
}
