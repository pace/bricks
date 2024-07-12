// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.

package util

import (
	"log"
	"net/http"
	"strings"
)

// configurableHandler is a wrapper for another middleware.
// It only calls the actual middleware if none of the ignoredPrefixes is prefix of the request path
type configurableHandler struct {
	ignoredPrefixes []string
	next            http.Handler
	actualHandler   http.Handler
}

// ConfigurableMiddlewareOption is a functional option to configure the handler
type ConfigurableMiddlewareOption func(*configurableHandler) error

// WithoutPrefixes allows to configure the ignoredPrefix slice
func WithoutPrefixes(prefix ...string) ConfigurableMiddlewareOption {
	return func(mdw *configurableHandler) error {
		mdw.ignoredPrefixes = append(mdw.ignoredPrefixes, prefix...)
		return nil
	}
}

// NewIgnorePrefixMiddleware creates a middleware that wraps the actualMiddleware. The handler of this middleware skips
// the actual middleware if the path has a prefix of the prefixes slice.
func NewIgnorePrefixMiddleware(actualMiddleware func(http.Handler) http.Handler, prefixes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return NewConfigurableHandler(next, actualMiddleware(next), WithoutPrefixes(prefixes...))
	}
}

// NewConfigurableHandler creates a configurableHandler,  that wraps anther handler.
// actualHandler is the handler, that is called if the request is not ignored
func NewConfigurableHandler(next, actualHandler http.Handler, cfgs ...ConfigurableMiddlewareOption) *configurableHandler {
	middleware := &configurableHandler{next: next, actualHandler: actualHandler}
	for _, cfg := range cfgs {
		if err := cfg(middleware); err != nil {
			log.Fatal(err)
		}
	}
	return middleware
}

// ServeHTTP tests if the path of the current request matches with any prefix of the list of ignored prefixes.
// If the Request should be ignored by the actual handler, the next handler is called, otherwise the actual handler is called
func (m configurableHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, prefix := range m.ignoredPrefixes {
		if strings.HasPrefix(r.URL.Path, prefix) {
			m.next.ServeHTTP(w, r)
			return
		}
	}
	m.actualHandler.ServeHTTP(w, r)
}
