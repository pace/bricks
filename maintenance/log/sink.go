package log

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// NoopSink is a Sink with no specific output.
// All logs passed to this Sink will be stored in
// it's internal storage but not passed further to
// the default logging output.
var NoopSink = &Sink{output: ioutil.Discard}

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
	jsonLogLines []string

	output  io.Writer
	rwmutex sync.RWMutex
}

// handlerWithSink returns a mux.MiddlewareFunc which
// adds a Sink to the request context. All logs
// corresponding to the request will be printed and stored
// in the Sink for later use.
func handlerWithSink() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sink := &Sink{}
			l := log.Ctx(r.Context()).Output(sink)
			ctx := ContextWithSink(l.WithContext(r.Context()), sink)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ToJSON returns a copy of the currently available
// logs in the Sink as json formatted []byte.
func (s *Sink) ToJSON() []byte {
	s.rwmutex.RLock()
	defer s.rwmutex.RUnlock()

	return []byte("[" + strings.Join(s.jsonLogLines, ",") + "]")
}

// Pretty returns the logs as string while using the
// zerolog.ConsoleWriter to format them in a human
// readable way
func (s *Sink) Pretty() string {
	buf := &bytes.Buffer{}
	writer := &zerolog.ConsoleWriter{
		Out:        buf,
		NoColor:    true,
		TimeFormat: "2006-01-02 15:04:05",
	}

	for _, str := range s.jsonLogLines {
		n, err := strings.NewReader(str).WriteTo(writer)
		if err != nil {
			log.Warn().Err(err).Msg("log.Sink.Pretty failed")
		} else if int(n) != len(str) {
			log.Warn().Msg("log.Sink.Pretty failed: could not return all logs")
		}
	}

	return buf.String()
}

// Write implements the io.Writer interface. This makes it
// possible to use the Sink as output in the zerolog.Output()
// func. Write stores all incoming logs in its internal store
// and calls Write() on the default output writer.
func (s *Sink) Write(b []byte) (int, error) {
	s.rwmutex.Lock()
	if s.output == nil {
		s.output = logOutput
	}

	s.jsonLogLines = append(s.jsonLogLines, string(b))
	s.rwmutex.Unlock()

	return s.output.Write(b)
}
