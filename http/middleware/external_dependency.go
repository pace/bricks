// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package middleware

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pace/bricks/maintenance/log"
)

// ExternalDependencyHeaderName name of the HTTP header that is used for reporting.
const ExternalDependencyHeaderName = "External-Dependencies"

// ExternalDependency middleware to report external dependencies.
func ExternalDependency(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var edc ExternalDependencyContext

		edw := externalDependencyWriter{
			ResponseWriter: w,
			edc:            &edc,
		}
		r = r.WithContext(ContextWithExternalDependency(r.Context(), &edc))
		next.ServeHTTP(&edw, r)
	})
}

func AddExternalDependency(ctx context.Context, name string, dur time.Duration) {
	ec := ExternalDependencyContextFromContext(ctx)
	if ec == nil {
		log.Ctx(ctx).Warn().Msgf("can't add external dependency %q with %s, because context is missing", name, dur)
		return
	}

	ec.AddDependency(name, dur)
}

type externalDependencyWriter struct {
	http.ResponseWriter
	header bool
	edc    *ExternalDependencyContext
}

// addHeader adds the external dependency header if not done already.
func (w *externalDependencyWriter) addHeader() {
	if !w.header {
		if len(w.edc.dependencies) > 0 {
			w.ResponseWriter.Header().Add(ExternalDependencyHeaderName, w.edc.String())
		}

		w.header = true
	}
}

func (w *externalDependencyWriter) Write(data []byte) (int, error) {
	w.addHeader()
	return w.ResponseWriter.Write(data)
}

func (w *externalDependencyWriter) WriteHeader(statusCode int) {
	w.addHeader()
	w.ResponseWriter.WriteHeader(statusCode)
}

// ContextWithExternalDependency creates a contex with the external provided dependencies.
func ContextWithExternalDependency(ctx context.Context, edc *ExternalDependencyContext) context.Context {
	return context.WithValue(ctx, (*ExternalDependencyContext)(nil), edc)
}

// ExternalDependencyContextFromContext returns the external dependencies context or nil.
func ExternalDependencyContextFromContext(ctx context.Context) *ExternalDependencyContext {
	if v := ctx.Value((*ExternalDependencyContext)(nil)); v != nil {
		out, ok := v.(*ExternalDependencyContext)
		if ok {
			return out
		}
	}

	return nil
}

// ExternalDependencyContext contains all dependencies that were seen
// during the request livecycle.
type ExternalDependencyContext struct {
	mu           sync.RWMutex
	dependencies map[string]time.Duration
}

func (c *ExternalDependencyContext) AddDependency(name string, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.dependencies == nil {
		c.dependencies = make(map[string]time.Duration)
	}

	c.dependencies[name] += duration
}

// String formats all external dependencies. The format is "name:duration[,name:duration]..." and sorted by name.
func (c *ExternalDependencyContext) String() string {
	var b strings.Builder

	for _, key := range slices.Sorted(maps.Keys(c.dependencies)) {
		b.WriteString(fmt.Sprintf("%s:%d,", key, c.dependencies[key].Milliseconds()))
	}

	return strings.TrimRight(b.String(), ",")
}

// Parse a external dependency value.
func (c *ExternalDependencyContext) Parse(s string) {
	for value := range strings.SplitSeq(s, ",") {
		index := strings.IndexByte(value, ':')
		if index == -1 {
			continue // ignore the invalid values
		}

		dur, err := strconv.ParseInt(value[index+1:], 10, 64)
		if err != nil {
			continue // ignore the invalid values
		}

		c.AddDependency(value[:index], time.Millisecond*time.Duration(dur))
	}
}
