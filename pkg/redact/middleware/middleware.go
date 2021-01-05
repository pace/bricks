// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/12/18 by Vincent Landgraf

package middleware

import (
	"net/http"

	"github.com/pace/bricks/pkg/redact"
)

// Redact provides a pattern redactor middleware to the request context
func Redact(next http.Handler) http.Handler {
	return RedactWithScheme(next, redact.Default)
}

// RedactWithScheme provides a pattern redactor middleware to the request context
// using the provided scheme
func RedactWithScheme(next http.Handler, redactor *redact.PatternRedactor) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := redactor.WithContext(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
