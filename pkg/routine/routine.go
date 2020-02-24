// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/01 by Marius Neugebauer

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

	"github.com/pace/bricks/http/oauth2"
	"github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/log"
)

// Run runs the given function in a new background context. The new context
// inherits the logger and oauth2 authentication of the parent context. Panics
// thrown in the function are logged and sent to sentry. The routines context is
// canceled if the program receives a shutdown signal (SIGINT, SIGTERM), if the
// returned CancelFunc is called, or if the routine returned.
func Run(parentCtx context.Context, routine func(context.Context)) (cancel context.CancelFunc) {
	// transfer logger, authentication and error info
	ctx := log.Ctx(parentCtx).WithContext(context.Background())
	ctx = oauth2.ContextTransfer(parentCtx, ctx)
	ctx = errors.ContextTransfer(parentCtx, ctx)
	// add routine number to logger
	num := atomic.AddInt64(&ctr, 1)
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
		routine(ctx)
		cancel()
	}()
	return
}

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
