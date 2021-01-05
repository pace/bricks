// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/12/18 by Vincent Landgraf

package redact

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func validatePattern(t *testing.T, p *regexp.Regexp, s string, expected bool) {
	assert.Equal(t, expected, p.MatchString(s), fmt.Sprintf("expected %q to match %v, but didn't", s, p))
}

func TestIBANPattern(t *testing.T) {
	validatePattern(t, PatternIBAN, "NL29INGB8731326943", true)
	validatePattern(t, PatternIBAN, "DE80500105172589468366", true)
	validatePattern(t, PatternIBAN, "NL29 INGB 8731 3269 43", true)
	validatePattern(t, PatternIBAN, "DE80 5001 0517 2589 4683 66", true)
	validatePattern(t, PatternIBAN, "CM7168156782527355483576522", true)
	validatePattern(t, PatternIBAN, "TL045597565817778146141", true)
	validatePattern(t, PatternIBAN, "AL85214511261456316638277339", true)
	validatePattern(t, PatternIBAN, "fL20-ING-B0-00-12-34-567", false)
	validatePattern(t, PatternIBAN, "fX22YYY1234567890123", true)
	validatePattern(t, PatternIBAN, "foo@i.ban", false)
}

func TestCreditCardPattern(t *testing.T) {
	// Testnumbers from https://www.paypalobjects.com/en_AU/vhelp/paypalmanager_help/credit_card_numbers.htm
	validatePattern(t, PatternCCAmericanExpress, "378282246310005", true)
	validatePattern(t, PatternCCAmericanExpress, "371449635398431", true)
	validatePattern(t, PatternCCAmericanExpress, "378734493671000", true)

	validatePattern(t, PatternCCDinersClub, "30569309025904", true)
	validatePattern(t, PatternCCDinersClub, "38520000023237", true)

	validatePattern(t, PatternCCDiscover, "6011111111111117", true)
	validatePattern(t, PatternCCDiscover, "6011000990139424", true)

	validatePattern(t, PatternCCJCB, "3530111333300000", true)
	validatePattern(t, PatternCCJCB, "3566002020360505", true)

	validatePattern(t, PatternCCMasterCard, "5555555555554444", true)
	validatePattern(t, PatternCCMasterCard, "5105105105105100", true)

	validatePattern(t, PatternCCVisa, "4111111111111111", true)
	validatePattern(t, PatternCCVisa, "4012888888881881", true)
	validatePattern(t, PatternCCVisa, "4222222222222", true)
}

func TestPatternJWT(t *testing.T) {
	validatePattern(t, PatternJWT, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c", true)
	validatePattern(t, PatternJWT, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJsb2dnZWRJbkFzIjoiYWRtaW4iLCJpYXQiOjE0MjI3Nzk2Mzh9.gzSraSYS8EXBxLN_oWnFSRgCzcmJmMjLiuyu5CSpyHI", true)
}
