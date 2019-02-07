// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/02/06 by Vincent Landgraf

package livetest

import (
	"context"
	"errors"
	"fmt"

	"lab.jamit.de/pace/go-microservice/maintenance/log"
)

// SkipNow is used as a panic if SkipNow is called on the test
var SkipNow = errors.New("skipped test")

// FailNow is used as a panic if FailNow is called on the test
var FailNow = errors.New("failed test")

type TestState string

var (
	StateRunning TestState = "running"
	StateOK      TestState = "ok"
	StateFailed  TestState = "failed"
	StateSkipped TestState = "skipped"
)

type T struct {
	name  string
	ctx   context.Context
	State TestState
}

func NewTestProxy(ctx context.Context, name string) *T {
	return &T{name: name, ctx: ctx, State: StateRunning}
}

func (t *T) Error(args ...interface{}) {
	log.Ctx(t.ctx).Error().Msg(fmt.Sprint(args...))
}

func (t *T) Errorf(format string, args ...interface{}) {
	log.Ctx(t.ctx).Error().Msgf(format, args...)
}

func (t *T) Fail() {
	log.Ctx(t.ctx).Info().Msg("Fail...")
	if t.State == StateRunning {
		t.State = StateFailed
	}
}

func (t *T) FailNow() {
	t.Fail()
	panic(FailNow)
}

func (t *T) Failed() bool {
	return t.State == StateFailed
}

func (t *T) Fatal(args ...interface{}) {
	log.Ctx(t.ctx).Error().Msg(fmt.Sprint(args...))
	t.FailNow()
}

func (t *T) Fatalf(format string, args ...interface{}) {
	log.Ctx(t.ctx).Error().Msgf(format, args...)
	t.FailNow()
}

func (t *T) Log(args ...interface{}) {
	log.Ctx(t.ctx).Info().Msg(fmt.Sprint(args...))
}

func (t *T) Logf(format string, args ...interface{}) {
	log.Ctx(t.ctx).Info().Msgf(format, args...)
}

func (t *T) Name() string {
	return t.name
}

func (t *T) Skip(args ...interface{}) {
	log.Ctx(t.ctx).Info().Msg("Skip...")
	log.Ctx(t.ctx).Info().Msg(fmt.Sprint(args...))
	if t.State == StateRunning {
		t.State = StateSkipped
	}
}

func (t *T) SkipNow() {
	log.Ctx(t.ctx).Info().Msg("Skip...")
	if t.State == StateRunning {
		t.State = StateSkipped
	}
	panic(SkipNow)
}

func (t *T) Skipf(format string, args ...interface{}) {
	log.Ctx(t.ctx).Info().Msg("Skip...")
	log.Ctx(t.ctx).Info().Msgf(format, args...)
	if t.State == StateRunning {
		t.State = StateSkipped
	}
}

func (t *T) Skipped() bool {
	return t.State == StateSkipped
}

func (t *T) okIfNoSkipFail() {
	if t.State == StateRunning {
		t.State = StateOK
	}
}
