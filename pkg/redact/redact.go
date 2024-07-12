// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package redact

import (
	"regexp"
)

type PatternRedactor struct {
	patterns []*regexp.Regexp
	scheme   RedactionScheme
}

// NewPatternRedactor creates a new redactor for masking certain patterns
func NewPatternRedactor(scheme RedactionScheme) *PatternRedactor {
	return &PatternRedactor{
		scheme: scheme,
	}
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
	r.patterns = append(r.patterns, patterns...)
}

// RemovePattern deletes a pattern from the redactor
func (r *PatternRedactor) RemovePattern(pattern *regexp.Regexp) {
	index := -1
	for i, p := range r.patterns {
		if p == pattern || p.String() == pattern.String() {
			index = i
			break
		}
	}
	if index >= 0 {
		r.patterns = append(r.patterns[:index], r.patterns[index+1:]...)
	}
}

func (r *PatternRedactor) SetScheme(scheme RedactionScheme) {
	r.scheme = scheme
}

func (r *PatternRedactor) Clone() *PatternRedactor {
	rc := NewPatternRedactor(r.scheme)
	rc.patterns = make([]*regexp.Regexp, len(r.patterns))
	copy(rc.patterns, r.patterns)
	return rc
}
