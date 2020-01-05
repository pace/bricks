// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/06 by Vincent Landgraf

package log

import (
	"context"
	"io"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/caarlos0/env"
	"github.com/rs/zerolog/hlog"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	isatty "github.com/mattn/go-isatty"
)

type config struct {
	LogLevel string `env:"LOG_LEVEL" envDefault:"debug"`
	Format   string `env:"LOG_FORMAT" envDefault:"auto"`
}

// map to translate the string log level
var levelMap = map[string]zerolog.Level{
	"debug":    zerolog.DebugLevel,
	"info":     zerolog.InfoLevel,
	"warn":     zerolog.WarnLevel,
	"error":    zerolog.ErrorLevel,
	"fatal":    zerolog.FatalLevel,
	"panic":    zerolog.PanicLevel,
	"disabled": zerolog.Disabled,
}

var cfg config

func init() {
	// parse log config
	err := env.Parse(&cfg)
	if err != nil {
		Fatalf("Failed to parse server environment: %v", err)
	}

	// translate log level
	v, ok := levelMap[strings.ToLower(cfg.LogLevel)]
	if !ok {
		Fatalf("Unknown log level: %q", cfg.LogLevel)
	}
	zerolog.SetGlobalLevel(v)
	log.Logger = log.Logger.Level(v)

	// use ico8601 (and UTC for json) as defined in https://lab.jamit.de/pace/web/meta/issues/11
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05"

	// auto detect log format
	if cfg.Format == "auto" {
		// if it is a terminal use the console writer
		if isatty.IsTerminal(os.Stdout.Fd()) {
			cfg.Format = "console"
		} else {
			cfg.Format = "json"
		}
	}

	switch cfg.Format {
	case "console":
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	case "json":
		// configure json output log
		log.Logger = log.Logger.Output(os.Stdout)
		zerolog.TimestampFunc = func() time.Time { return time.Now().UTC() }
	}
}

// RequestID returns a unique request id or an empty string if there is none
func RequestID(r *http.Request) string {
	id, ok := hlog.IDFromRequest(r)
	if ok {
		return id.String()
	}
	return ""
}

// RequestIDFromContext returns a unique request id or an empty string if there is none
func RequestIDFromContext(ctx context.Context) string {
	// create dummy request to get the request id
	r, err := http.NewRequest("GET", "", nil)
	if err != nil {
		panic(err)
	}

	return RequestID(r.WithContext(ctx))
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

// Stack prints the stack of the calling goroutine
func Stack(ctx context.Context) {
	for _, line := range strings.Split(string(debug.Stack()), "\n") {
		if line != "" {
			Ctx(ctx).Error().Msg(line)
		}
	}
}

// WithContext returns context with enabled logger.
// This overwrites a logger that is set on the context already
// use this if you are not inside a request context.
func WithContext(ctx context.Context) context.Context {
	return log.Logger.WithContext(ctx)
}

// Output duplicates the current logger and sets w as its output.
func Output(w io.Writer) *zerolog.Logger {
	logger := Logger().Output(w)
	return &logger
}
