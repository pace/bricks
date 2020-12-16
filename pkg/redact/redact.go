// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/12/16 by Florian S.

package redact

import (
	"context"
	"regexp"
)

type RedactionScheme func(string) string

type patternRedactorKey struct{}

type PatternRedactor struct {
	patterns map[string]*regexp.Regexp
	scheme   RedactionScheme
}

func (r *PatternRedactor) Mask(data string) string {
	for _, pattern := range r.patterns {
		if pattern == nil {
			continue
		}
		data = pattern.ReplaceAllStringFunc(data, r.scheme)
	}
	return data
}

// AddPattern adds patterns to the redactor
func (r *PatternRedactor) AddPatterns(patterns ...*regexp.Regexp) {
	for _, pattern := range patterns {
		r.patterns[pattern.String()] = pattern
	}
}

// RemovePattern deletes a pattern from the redactor
func (r *PatternRedactor) RemovePattern(pattern *regexp.Regexp) {
	delete(r.patterns, pattern.String())
}

func (r *PatternRedactor) SetScheme(scheme RedactionScheme) {
	r.scheme = scheme
}

// NewPatternRedactor creates a new redactor for masking certain patterns
func NewPatternRedactor(scheme RedactionScheme) *PatternRedactor {
	patterns := make(map[string]*regexp.Regexp)
	return &PatternRedactor{
		patterns: patterns,
		scheme:   scheme,
	}
}

func (r *PatternRedactor) Clone() *PatternRedactor {
	rc := NewPatternRedactor(r.scheme)
	for k, v := range r.patterns {
		rc.patterns[k] = regexp.MustCompile(v.String())
	}
	return rc
}

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

func RedactionSchemeDoNothing() func(string) string {
	return func(old string) string {
		return old
	}
}

// RedactionSchemeKeepLast replaces all runes in the string with an asterisk
// except the last NUM runes
func RedactionSchemeKeepLast(num int) func(string) string {
	return func(old string) string {
		runes := []rune(old)
		for i := 0; i < len(runes)-num; i++ {
			runes[i] = '*'
		}
		return string(runes)
	}
}
