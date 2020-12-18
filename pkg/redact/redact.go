// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/12/16 by Florian S.

package redact

import (
	"regexp"
)

type PatternRedactor struct {
	patterns map[string]*regexp.Regexp
	scheme   RedactionScheme
}

// NewPatternRedactor creates a new redactor for masking certain patterns
func NewPatternRedactor(scheme RedactionScheme) *PatternRedactor {
	patterns := make(map[string]*regexp.Regexp)
	return &PatternRedactor{
		patterns: patterns,
		scheme:   scheme,
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

func (r *PatternRedactor) Clone() *PatternRedactor {
	rc := NewPatternRedactor(r.scheme)
	for k, v := range r.patterns {
		rc.patterns[k] = regexp.MustCompile(v.String())
	}
	return rc
}
