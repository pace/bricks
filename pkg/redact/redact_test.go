// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package redact_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pace/bricks/pkg/redact"
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

func TestRedactionSchemeJWT(t *testing.T) {
	redactor := redact.NewPatternRedactor(redact.RedactionSchemeKeepLastJWTNoSignature(4))
	redactor.AddPatterns(redact.PatternJWT)
	redactor.AddPatterns(redact.PatternIBAN)
	redactor.AddPatterns(regexp.MustCompile("AnotherSpecialSecret"))

	originalString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
	expectedString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.***************************************sw5c"

	res := redactor.Mask(originalString)
	assert.Equal(t, expectedString, res)

	originalString = `Here we have an IBAN: DE12345678909876543210
and a SuperSecretSpecialString, as well as AnotherSpecialSecret`
	expectedString = `Here we have an IBAN: ******************3210
and a SuperSecretSpecialString, as well as ****************cret`

	res = redactor.Mask(originalString)
	assert.Equal(t, expectedString, res)
}
