// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/12/16 by Florian S.

package redact_test

import (
	"regexp"
	"testing"

	"github.com/pace/bricks/pkg/redact"

	"github.com/stretchr/testify/assert"
)

func TestRedactionSchemeKeepLast(t *testing.T) {
	patterns := []*regexp.Regexp{
		regexp.MustCompile("DE12345678909876543210"),
		regexp.MustCompile("SuperSecretSpecialString"),
	}
	redactor := redact.NewPatternRedactor(redact.RedactionSchemeKeepLast(4))
	redactor.AddPatterns(patterns...)
	redactor.AddPatterns(regexp.MustCompile("AnotherSpecialSecret"))

	originalString := `Here we have an IBAN: DE12345678909876543210
and a SuperSecretSpecialString, as well as AnotherSpecialSecret`
	expectedString1 := `Here we have an IBAN: ******************3210
and a ********************ring, as well as ****************cret`
	expectedString2 := `Here we have an IBAN: DE12345678909876543210
and a ********************ring, as well as ****************cret`

	res := redactor.Mask(originalString)
	assert.Equal(t, expectedString1, res)
	redactor.RemovePattern(regexp.MustCompile("DE12345678909876543210"))
	res = redactor.Mask(originalString)
	assert.Equal(t, expectedString2, res)
}
