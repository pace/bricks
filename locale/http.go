// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/04/16 by Vincent Landgraf

package locale

import "net/http"

const (
	HeaderAcceptLanguage = "Accept-Language"
	HeaderAcceptTimezone = "Accept-Timezone"
)

// FromRequest creates a locale based on the accept headers from the given request.
func FromRequest(r *http.Request) *Locale {
	return NewLocale(r.Header.Get(HeaderAcceptLanguage), r.Header.Get(HeaderAcceptTimezone))
}

// Request returns the passed request with added accept headers.
// The request is returned for convenience.
func (l Locale) Request(r *http.Request) *http.Request {
	if l.HasLanguage() {
		r.Header.Set(HeaderAcceptLanguage, l.acceptLanguage)
	}
	if l.HasTimezone() {
		r.Header.Set(HeaderAcceptTimezone, l.acceptTimezone)
	}
	return r
}

// Middleware takes the accept lang and timezone info and
// stores them in the context
type Middleware struct {
	next http.Handler
}

// ServeHTTP adds the locale to the request context
func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := FromRequest(r)
	r = r.WithContext(WithLocale(r.Context(), l))
	m.next.ServeHTTP(w, r)
}

// Handler builds new Middleware
func Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return &Middleware{next: next}
	}
}
