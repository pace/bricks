// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/12/17 by Charlotte Pröller

package util

import (
	"log"
	"net/http"
	"strings"

)

// ConfigurableMiddleware is a wrapper for another middleware.
// It only calls the actual middleware if none of the ignoredPrefixes is prefix of the request path
type ConfigurableMiddleware struct {
	ignoredPrefixes  []string
	next             http.Handler
	actualMiddleware http.Handler
}

// ConfigurableMiddlewareOption is ap functional Option to configure the middleware
type ConfigurableMiddlewareOption func(*ConfigurableMiddleware) error

// WithoutPrefixes allows to configure the ignoredPrefix slice
func WithoutPrefixes(prefix ...string) ConfigurableMiddlewareOption {
	return func(mdw *ConfigurableMiddleware) error {
		mdw.ignoredPrefixes = append(mdw.ignoredPrefixes, prefix...)
		return nil
	}
}

// NewIgnorePrefixMiddleware creates a ConfigurableMiddleware
// actualMiddleware is the actual middleware, that is called if the request is not ignored
func NewIgnorePrefixMiddleware(next, actualMiddleware http.Handler, cfgs ...ConfigurableMiddlewareOption) *ConfigurableMiddleware {
	middleware :=  &ConfigurableMiddleware{next: next, actualMiddleware: actualMiddleware}
	for _, cfg := range cfgs {
		if err := cfg(middleware); err != nil {
			log.Fatal(err)
		}
	}
	return middleware
}

// ServeHTTP tests if the path of the current request matches with any prefix of the list of ignored prefixes.
// If the Request should be ignored by the actual middleware, the next handler is called, otherwise the actual middleware is called
func (m ConfigurableMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, prefix := range m.ignoredPrefixes {
		if strings.HasPrefix(r.URL.Path, prefix) {
			m.next.ServeHTTP(w, r)
			return
		}
	}
	m.actualMiddleware.ServeHTTP(w, r)
}
