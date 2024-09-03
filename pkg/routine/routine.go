// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.

// Package routine helps in starting background tasks.
package routine

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/getsentry/sentry-go"

	"github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/log"
	pkgcontext "github.com/pace/bricks/pkg/context"
)

type options struct {
	keepRunningOneInstance bool
}

// Option specifies how a routine is run.
type Option func(*options)

// Ideas for options in the future:
//  * Timeout/Deadline for the context
//  * Cronjob("30 */2 * * *"): run according to the crontab notation
//  * AutoRestart(time.Duration): restart after routine finishes, but at most
//    once per duration
//  * Every(time.Duration): run regularly, use redis to get consistent behaviour
//    on process restarts
//  * Workers(int): run the routine a number of times in parallel
//  * Delay(time.Duration): run once after timeout
//  * allow useful combinations of OnceSimultaneously and other options
//  * join a group explicitly by selecting a different redis database
//  * join a group explicitly by choosing a different name

// KeepRunningOneInstance returns an option that runs the routine once
// simultaneously and keeps it running, restarting it if necessary. Subsequent
// calls of the routine are stalled until the previous call returned.
//
// In clusters with multiple processes only one caller actually executes the
// routine, but it can be any one of the callers. If the routine finishes any
// caller including the same caller can execute it. If the process of the caller
// executing the routine is stopped another caller executes the routine. If the
// last caller is stopped, the routine stops executing due to lack of callers
// that can continue execution. There is no automatic takeover of the routine
// by other running members of the group, that did not call it explicitly.
//
// Due to lack of a better name, the name of this option is quite verbose. Feel
// free to propose any better name as an alias for this option.
func KeepRunningOneInstance() Option {
	return func(o *options) {
		o.keepRunningOneInstance = true
	}
}

// RunNamed runs a routine like Run does. Additionally it assigns the routine a
// name and allows using options to control how the routine is run. Routines
// with the same name show consistent behaviour for the options, like mutual
// exclusion, across their group. By default all callers that share the same
// redis database are members of the same group, no matter whether they are
// goroutines of a single process or of processes running on different hosts.
// The default redis database is configured via the REDIS_* environment
// variables.
func RunNamed(parentCtx context.Context, name string, routine func(context.Context), opts ...Option) (cancel context.CancelFunc) {
	o := options{}
	for _, opt := range opts {
		opt(&o)
	}

	if o.keepRunningOneInstance {
		routine = (&routineThatKeepsRunningOneInstance{
			Name:    name,
			Routine: routine,
		}).Run
	}

	return Run(parentCtx, routine)
}

// Run runs the given function in a new background context. The new context
// inherits the logger and oauth2 authentication of the parent context. Panics
// thrown in the function are logged and sent to sentry. The routines context is
// canceled if the program receives a shutdown signal (SIGINT, SIGTERM), if the
// returned CancelFunc is called, or if the routine returned.
func Run(parentCtx context.Context, routine func(context.Context)) (cancel context.CancelFunc) {
	ctx := pkgcontext.Transfer(parentCtx)

	// add routine number to context and logger
	num := atomic.AddInt64(&ctr, 1)

	span := sentry.StartTransaction(ctx, fmt.Sprintf("Routine %d", num), sentry.WithOpName("function"))
	defer span.Finish()

	ctx = span.Context()

	ctx = context.WithValue(ctx, ctxNumKey{}, num)
	logger := log.Ctx(ctx).With().Int64("routine", num).Logger()
	ctx = logger.WithContext(ctx)
	// get cancel function
	ctx, cancel = context.WithCancel(ctx)
	// register context to be cancelled when the program is shut down
	contextsMx.Lock()
	contexts[num] = cancel
	contextsMx.Unlock()
	// deregister the above if context is done
	go func() {
		<-ctx.Done()
		// In case of a shutdown, this will block forever. But it doesn't hurt,
		// because the program is about to exit anyway.
		contextsMx.Lock()
		defer contextsMx.Unlock()
		delete(contexts, num)
	}()
	go func() {
		defer errors.HandleWithCtx(ctx, fmt.Sprintf("routine %d", num)) // handle panics
		defer cancel()
		routine(ctx)
	}()
	return
}

type ctxNumKey struct{}

var (
	contextsMx sync.Mutex
	contexts   = map[int64]context.CancelFunc{}
	ctr        int64
)

// Starts a go routine that cancels all contexts for routines created by Run if
// we receive a SIGINT/SIGTERM. This allows those routines to gracefully handle
// the shutdown.
func init() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c // block until SIGINT/SIGTERM is received
		signal.Stop(c)
		// no unlock, to block creating new routines while the program exits
		contextsMx.Lock()
		// Cancel all contexts. For contexts that are already done this is a
		// no-op.
		log.Logger().Info().
			Int("count", len(contexts)).
			Ints64("routines", routineNumbers()).
			Msg("received shutdown signal, canceling all running routines")
		for _, cancel := range contexts {
			cancel()
		}
	}()
}

func routineNumbers() []int64 {
	routines := make([]int64, 0, len(contexts))
	for num := range contexts {
		routines = append(routines, num)
	}
	return routines
}
