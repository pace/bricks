package context

import (
	"context"

	"github.com/getsentry/sentry-go"
	http "github.com/pace/bricks/http/middleware"
	"github.com/pace/bricks/http/oauth2"
	"github.com/pace/bricks/locale"
	"github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/log"
	"github.com/pace/bricks/maintenance/log/hlog"
	"github.com/pace/bricks/pkg/redact"
	"github.com/pace/bricks/pkg/tracking/utm"
)

// Transfer takes the logger, log.Sink, authentication, request and
// error info from the given context and returns a complete
// new context with all these objects.
func Transfer(in context.Context) context.Context {
	// transfer logger, log.Sink, authentication and error info
	out := TransferTracingContext(in, context.Background())
	out = log.Ctx(in).WithContext(out)
	out = log.SinkContextTransfer(in, out)
	out = oauth2.ContextTransfer(in, out)
	out = errors.ContextTransfer(in, out)
	out = http.ContextTransfer(in, out)
	out = redact.ContextTransfer(in, out)
	out = utm.ContextTransfer(in, out)
	out = hlog.ContextTransfer(in, out)
	out = locale.ContextTransfer(in, out)
	out = TransferExternalDependencyContext(in, out)

	return out
}

func TransferTracingContext(in, out context.Context) context.Context {
	span := sentry.SpanFromContext(in)
	if span == nil {
		return out
	}

	return span.Context()
}

func TransferExternalDependencyContext(in, out context.Context) context.Context {
	edc := http.ExternalDependencyContextFromContext(in)
	if edc == nil {
		return out
	}
	return http.ContextWithExternalDependency(out, edc)
}
