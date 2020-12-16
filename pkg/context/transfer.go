package context

import (
	"context"

	"github.com/pace/bricks/http"
	"github.com/pace/bricks/http/oauth2"
	"github.com/pace/bricks/locale"
	"github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/log"
	"github.com/pace/bricks/pkg/redact"
)

// Transfer takes the logger, log.Sink, authentication, request and
// error info from the given context and returns a complete
// new context with all these objects.
func Transfer(in context.Context) context.Context {
	// transfer logger, log.Sink, authentication and error info
	out := log.Ctx(in).WithContext(context.Background())
	out = log.SinkContextTransfer(in, out)
	out = oauth2.ContextTransfer(in, out)
	out = errors.ContextTransfer(in, out)
	out = http.ContextTransfer(in, out)
	out = redact.ContextTransfer(in, out)
	return locale.ContextTransfer(in, out)
}
