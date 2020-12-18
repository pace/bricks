// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/12/18 by Vincent Landgraf

package middleware

import (
	"net/http"

	"github.com/pace/bricks/pkg/redact"
)

// redactionSafe last 4 digits are usually concidered safe (e.g. credit cards, iban, ...)
const redactionSafe = 4

// Redact provides a pattern redactor middleware to the request context
func Redact(next http.Handler) http.Handler {
	return RedactWithScheme(next, redact.RedactionSchemeKeepLast(redactionSafe))
}

// RedactWithScheme provides a pattern redactor middleware to the request context
// using the provided scheme
func RedactWithScheme(next http.Handler, scheme redact.RedactionScheme) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var pr redact.PatternRedactor
		pr.SetScheme(scheme)
		ctx := pr.WithContext(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
