package sentry

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/getsentry/sentry-go"
)

func init() {
	var tracesSampleRate float64 = 0.1

	val := strings.TrimSpace(os.Getenv("SENTRY_TRACES_SAMPLE_RATE"))
	if val != "" {
		var err error

		tracesSampleRate, err = strconv.ParseFloat(val, 64)
		if err != nil {
			log.Fatalf("failed to parse SENTRY_TRACES_SAMPLE_RATE: %v", err)
		}
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		Environment:      os.Getenv("ENVIRONMENT"),
		EnableTracing:    true,
		TracesSampleRate: tracesSampleRate,
		BeforeSendTransaction: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			// Drop request body.
			if event.Request != nil {
				event.Request.Data = ""
			}

			return event
		},
	})
	if err != nil {
		log.Fatalf("sentry.Init: %v", err)
	}
}
