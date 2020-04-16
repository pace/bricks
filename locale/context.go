// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/04/16 by Vincent Landgraf

package locale

import "context"

// ctx private key type to seal the access
type ctx string

// tokenKey private key to seal the access
var tokenKey = ctx("locale")

// WithLocale creates a new context with the passed locale
func WithLocale(ctx context.Context, locale *Locale) context.Context {
	return context.WithValue(ctx, tokenKey, locale)
}

// FromCtx returns the locale from the context.
// The returned locale is always not nil.
func FromCtx(ctx context.Context) (*Locale, bool) {
	val := ctx.Value(tokenKey)
	if val == nil {
		return new(Locale), false
	}
	l, ok := val.(*Locale)
	return l, ok
}

// ContextTransfer sources the locale from the sourceCtx
// and returns a new context based on the targetCtx
func ContextTransfer(sourceCtx context.Context, targetCtx context.Context) context.Context {
	l, _ := FromCtx(sourceCtx)
	return WithLocale(targetCtx, l)
}
