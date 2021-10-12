// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/12/18 by Vincent Landgraf

package redact

import "regexp"

// Sources:
// CreditCard: https://www.regular-expressions.info/creditcard.html

// AllPatterns is a list of all default redaction patterns
var AllPatterns = []*regexp.Regexp{
	PatternIBAN,
	PatternJWT,
	PatternCCVisa,
	PatternCCMasterCard,
	PatternCCAmericanExpress,
	PatternCCDinersClub,
	PatternCCDiscover,
	PatternCCJCB,
	PatternBasicAuthBase64,
}

var (
	PatternIBAN = regexp.MustCompile(
		`[a-zA-Z]{2}` + // DE, NL, ...
			`[0-9]{2}` + // 80
			`(?:[ ]?[0-9a-zA-Z]{4})` + // 5001, INGB
			`(?:[ ]?[0-9]{4}){2,3}` + // 0517 2589 4683, 8731 3269
			`(?:[ ]?[0-9]{1,2})?`, // 43, 66
	)

	// All Visa card numbers start with a 4. New cards have 16 digits. Old cards have 13.
	PatternCCVisa = regexp.MustCompile(`4[0-9]{12}(?:[0-9]{3})?`)

	// MasterCard numbers either start with the numbers 51 through 55 or with the numbers 2221 through 2720. All have 16 digits.
	PatternCCMasterCard = regexp.MustCompile(`(?:5[1-5][0-9]{2}|222[1-9]|22[3-9][0-9]|2[3-6][0-9]{2}|27[01][0-9]|2720)[0-9]{12}`)

	// American Express card numbers start with 34 or 37 and have 15 digits.
	PatternCCAmericanExpress = regexp.MustCompile(`3[47][0-9]{13}`)

	// Diners Club card numbers begin with 300 through 305, 36 or 38. All have 14 digits. There are Diners Club cards that begin with 5 and have 16 digits. These are a joint venture between Diners Club and MasterCard, and should be processed like a MasterCard.
	PatternCCDinersClub = regexp.MustCompile(`3(?:0[0-5]|[68][0-9])[0-9]{11}`)

	// Discover card numbers begin with 6011 or 65. All have 16 digits.
	PatternCCDiscover = regexp.MustCompile(`6(?:011|5[0-9]{2})[0-9]{12}`)

	// JCB cards beginning with 2131 or 1800 have 15 digits. JCB cards beginning with 35 have 16 digits.
	PatternCCJCB = regexp.MustCompile(`(?:2131|1800|35\d{3})\d{11}`)

	// PatternJWT JsonWebToken
	PatternJWT = regexp.MustCompile(`(?:ey[a-zA-Z0-9=_-]+\.){2}[a-zA-Z0-9=_-]+`)

	//PatternBasicAuthBase match any: Basic YW55IGNhcm5hbCBwbGVhcw== does not validate base64 string
	PatternBasicAuthBase64 = regexp.MustCompile(`Basic ([a-zA-Z0-9=]*)`)
)
