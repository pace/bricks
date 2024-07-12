// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package redact

import "context"

type patternRedactorKey struct{}

// WithContext allows storing the PatternRedactor inside a context for passing it on
func (r *PatternRedactor) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, patternRedactorKey{}, r)
}

// Ctx returns the PatternRedactor stored within the context. If no redactor
// has been defined, an empty redactor is returned that does nothing
func Ctx(ctx context.Context) *PatternRedactor {
	if rd, ok := ctx.Value(patternRedactorKey{}).(*PatternRedactor); ok {
		return rd.Clone()
	}
	return NewPatternRedactor(RedactionSchemeDoNothing())
}

// ContextTransfer copies a request representation from one context to another.
func ContextTransfer(ctx, targetCtx context.Context) context.Context {
	if redactor := Ctx(ctx); redactor != nil {
		return context.WithValue(targetCtx, patternRedactorKey{}, redactor)
	}
	return targetCtx
}
