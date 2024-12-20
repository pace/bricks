package sentry

import (
	"log"
	"os"

	"github.com/getsentry/sentry-go"
)

func init() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		Environment:      os.Getenv("ENVIRONMENT"),
		EnableTracing:    true,
		TracesSampleRate: 1.0,
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
