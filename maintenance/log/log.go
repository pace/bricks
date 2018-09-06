// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/06 by Vincent Landgraf

package log

import (
	"context"
	"net/http"
	"os"

	"github.com/rs/zerolog/hlog"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	isatty "github.com/mattn/go-isatty"
)

func init() {
	// if it is a terminal use the console writer
	if isatty.IsTerminal(os.Stdout.Fd()) {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
		return
	}
	log.Logger = log.Logger.Output(os.Stdout)
}

// Req returns the logger for the passed request
func Req(r *http.Request) *zerolog.Logger {
	return hlog.FromRequest(r)
}

// Ctx returns the logger for the passed context
func Ctx(ctx context.Context) *zerolog.Logger {
	return log.Ctx(ctx)
}

// Logger returns the current logger instance
func Logger() *zerolog.Logger {
	return &log.Logger
}
