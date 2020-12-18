// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/12/14 by Vincent Landgraf

package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pace/bricks/maintenance/log"
)

// depFormat is the format of a single dependency report
const depFormat = "%s:%d"

// ExternalDependencyHeaderName name of the HTTP header that is used for reporting
const ExternalDependencyHeaderName = "External-Dependencies"

// ExternalDependency middleware to report external dependencies
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

// addHeader adds the external dependency header if not done already
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

// ContextWithExternalDependency creates a contex with the external provided dependencies
func ContextWithExternalDependency(ctx context.Context, edc *ExternalDependencyContext) context.Context {
	return context.WithValue(ctx, (*ExternalDependencyContext)(nil), edc)
}

// ExternalDependencyContextFromContext returns the external dependencies context or nil
func ExternalDependencyContextFromContext(ctx context.Context) *ExternalDependencyContext {
	if v := ctx.Value((*ExternalDependencyContext)(nil)); v != nil {
		return v.(*ExternalDependencyContext)
	}
	return nil
}

// ExternalDependencyContext contains all dependencies that were seen
// during the request livecycle
type ExternalDependencyContext struct {
	mu           sync.RWMutex
	dependencies []externalDependency
}

func (c *ExternalDependencyContext) AddDependency(name string, duration time.Duration) {
	c.mu.Lock()
	c.dependencies = append(c.dependencies, externalDependency{
		Name:     name,
		Duration: duration,
	})
	c.mu.Unlock()
}

// String formats all external dependencies
func (c *ExternalDependencyContext) String() string {
	var b strings.Builder
	sep := len(c.dependencies) - 1
	for _, dep := range c.dependencies {
		b.WriteString(dep.String())
		if sep > 0 {
			b.WriteByte(',')
			sep--
		}
	}
	return b.String()
}

// Parse a external dependency value
func (c *ExternalDependencyContext) Parse(s string) {
	values := strings.Split(s, ",")
	for _, value := range values {
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

// externalDependency represents one external dependency that
// was involved in the process to creating a response
type externalDependency struct {
	Name     string        // canonical name of the source
	Duration time.Duration // time spend with the external dependency
}

// String returns a formated single external dependency
func (r externalDependency) String() string {
	return fmt.Sprintf(depFormat, r.Name, r.Duration.Milliseconds())
}
