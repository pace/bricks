// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/02/06 by Vincent Landgraf

package livetest

import (
	"context"
	"errors"
	"fmt"

	"github.com/pace/bricks/maintenance/log"
)

// ErrSkipNow is used as a panic if ErrSkipNow is called on the test
var ErrSkipNow = errors.New("skipped test")

// ErrFailNow is used as a panic if ErrFailNow is called on the test
var ErrFailNow = errors.New("failed test")

// TestState represents the state of a test
type TestState string

var (
	// StateRunning first state
	StateRunning TestState = "running"
	// StateOK test was executed without failure
	StateOK TestState = "ok"
	// StateFailed test was executed with failure
	StateFailed TestState = "failed"
	// StateSkipped test was skipped
	StateSkipped TestState = "skipped"
)

// T implements a similar interface than testing.T
type T struct {
	name  string
	ctx   context.Context
	state TestState
}

// NewTestProxy creates a new text proxy using the given context
// and name.
func NewTestProxy(ctx context.Context, name string) *T {
	return &T{name: name, ctx: ctx, state: StateRunning}
}

// Context returns the livetest context. Useful
// for passing timeout and/or logging constraints from
// the test executor to the individual case
func (t *T) Context() context.Context {
	return t.ctx
}

// Error logs an error message with the test
func (t *T) Error(args ...interface{}) {
	log.Ctx(t.ctx).Error().Msg(fmt.Sprint(args...))
	t.Fail()
}

// Errorf logs an error message with the test
func (t *T) Errorf(format string, args ...interface{}) {
	log.Ctx(t.ctx).Error().Msgf(format, args...)
	t.Fail()
}

// Fail marks the test as failed
func (t *T) Fail() {
	log.Ctx(t.ctx).Info().Msg("Fail...")
	if t.state == StateRunning {
		t.state = StateFailed
	}
}

// FailNow marks the test as failed and skips further execution
func (t *T) FailNow() {
	t.Fail()
	panic(ErrFailNow)
}

// Failed returns true if the test was marked as failed
func (t *T) Failed() bool {
	return t.state == StateFailed
}

// Fatal logs the passed message in the context of the test and fails the test
func (t *T) Fatal(args ...interface{}) {
	log.Ctx(t.ctx).Error().Msg(fmt.Sprint(args...))
	t.FailNow()
}

// Fatalf logs the passed message in the context of the test and fails the test
func (t *T) Fatalf(format string, args ...interface{}) {
	log.Ctx(t.ctx).Error().Msgf(format, args...)
	t.FailNow()
}

// Log logs the passed message in the context of the test
func (t *T) Log(args ...interface{}) {
	log.Ctx(t.ctx).Info().Msg(fmt.Sprint(args...))
}

// Logf logs the passed message in the context of the test
func (t *T) Logf(format string, args ...interface{}) {
	log.Ctx(t.ctx).Info().Msgf(format, args...)
}

// Name returns the name of the test
func (t *T) Name() string {
	return t.name
}

// Skip logs reason and marks the test as skipped
func (t *T) Skip(args ...interface{}) {
	log.Ctx(t.ctx).Info().Msg("Skip...")
	log.Ctx(t.ctx).Info().Msg(fmt.Sprint(args...))
	if t.state == StateRunning {
		t.state = StateSkipped
	}
}

// SkipNow skips the test immediately
func (t *T) SkipNow() {
	log.Ctx(t.ctx).Info().Msg("Skip...")
	if t.state == StateRunning {
		t.state = StateSkipped
	}
	panic(ErrSkipNow)
}

// Skipf marks the test as skippend and log a reason
func (t *T) Skipf(format string, args ...interface{}) {
	log.Ctx(t.ctx).Info().Msg("Skip...")
	log.Ctx(t.ctx).Info().Msgf(format, args...)
	if t.state == StateRunning {
		t.state = StateSkipped
	}
}

// Skipped returns true if the test was skipped
func (t *T) Skipped() bool {
	return t.state == StateSkipped
}

func (t *T) okIfNoSkipFail() {
	if t.state == StateRunning {
		t.state = StateOK
	}
}
