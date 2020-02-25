package log

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/mux"
)

type sinkKey struct{}

// ContextWithSink wraps the given context in a new context with
// the given Sink stored as value.
func ContextWithSink(ctx context.Context, sink *Sink) context.Context {
	return context.WithValue(ctx, sinkKey{}, sink)
}

// SinkFromContext returns the Sink of the given context if it exists
func SinkFromContext(ctx context.Context) (*Sink, bool) {
	sink, ok := ctx.Value(sinkKey{}).(*Sink)
	return sink, ok
}

// Sink respresents a log sink which is used to store
// logs, created with log.Ctx(ctx), inside the context
// and use them at a later point in time
type Sink struct {
	logs []string

	rwmutex sync.RWMutex
}

// handlerWithSink returns a mux.MiddlewareFunc which
// adds the Sink to the request context. All logs
// corresponding to the request will be printed and stored
// in the Sink for later use.
func handlerWithSink(sink *Sink) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(ContextWithSink(r.Context(), sink)))
		})
	}
}

// ToJSON returns a copy of the currently available
// logs in the Sink as json.RawMessage.
func (s *Sink) ToJSON() json.RawMessage {
	s.rwmutex.RLock()
	defer s.rwmutex.RUnlock()

	return []byte("[" + strings.Join(s.logs, ",") + "]")
}

// Write implements the io.Writer interface. This makes it
// possible to use the Sink as output in the zerolog.Output()
// func. Write stores all incoming logs in its internal store
// and calls Write() on the default output writer.
func (s *Sink) Write(b []byte) (int, error) {
	s.rwmutex.Lock()
	s.logs = append(s.logs, string(b))
	s.rwmutex.Unlock()

	return logOutput.Write(b)
}
