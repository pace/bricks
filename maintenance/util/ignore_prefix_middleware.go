// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/12/17 by Charlotte Pröller

package util

import (
	"net/http"
	"strings"
)

// IgnorePrefixMiddleware is a wrapper for another middleware.
// It only calls the actual middleware if none of the ignoredPrefixes is prefix of the request path
type IgnorePrefixMiddleware struct {
	ignoredPrefixes  []string
	next             http.Handler
	actualMiddleware http.Handler
}

// NewIgnorePrefixMiddleware creates a IgnorePrefixMiddleware
// actualMiddleware is the actual middleware, that is called if the request is not ignored
func NewIgnorePrefixMiddleware(next, actualMiddleware http.Handler, ignoredPrefixes ...string) *IgnorePrefixMiddleware {
	return &IgnorePrefixMiddleware{ignoredPrefixes: ignoredPrefixes, next: next, actualMiddleware: actualMiddleware}
}

// ServeHTTP tests if the path of the current request matches with any prefix of the list of ignored prefixes.
// If the Request should be ignored by the actual middleware, the next handler is called, otherwise the actual middleware is called
func (m IgnorePrefixMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, prefix := range m.ignoredPrefixes {
		if strings.HasPrefix(r.URL.Path, prefix) {
			m.next.ServeHTTP(w, r)
			return
		}
	}
	m.actualMiddleware.ServeHTTP(w, r)
}
