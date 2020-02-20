// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/01 by Marius Neugebauer

// Package routine helps in starting background tasks.
package routine

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/pace/bricks/http/oauth2"
	"github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/log"
)

// Forever is so far in the future that it basically never ends.
var Forever = time.Duration(1<<63 - 1)

// Run runs the given function in a new context with the given timeout. The new
// context inherits the logger and oauth2 authentication of the passed context.
// Panics thrown in the function are logged and sent to sentry. The routines
// context is cancelled if the program receives a shutdown signal (SIGINT,
// SIGTERM).
func Run(ctx context.Context, timeout time.Duration, routine func(context.Context)) (cancel context.CancelFunc) {
	// transfer logger, authentication and error info
	routineCtx := log.Ctx(ctx).WithContext(context.Background())
	routineCtx = oauth2.ContextTransfer(ctx, routineCtx)
	routineCtx = errors.ContextTransfer(ctx, routineCtx)
	// apply timeout
	routineCtx, cancel = context.WithTimeout(routineCtx, timeout)
	// register context to be cancelled when the program is shut down
	contextsMx.Lock()
	contexts[routineCtx] = cancel
	contextsMx.Unlock()
	// deregister the above if context is done
	go func() {
		<-routineCtx.Done()
		// In case of a shutdown, this will block forever. But it doesn't hurt,
		// because the program is about to exit anyway.
		contextsMx.Lock()
		defer contextsMx.Unlock()
		delete(contexts, routineCtx)
	}()
	go func() {
		defer errors.HandleWithCtx(routineCtx, "routine") // handle panics
		routine(routineCtx)
		cancel()
	}()
	return
}

var (
	contextsMx sync.Mutex
	contexts   = map[context.Context]context.CancelFunc{}
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
		log.Logger().Info().Int("count", len(contexts)).Msg("received shutdown signal, cancelling all running routines")
		for _, cancel := range contexts {
			cancel()
		}
	}()
}
