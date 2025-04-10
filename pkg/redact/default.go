// Copyright © 2021 by PACE Telematics GmbH. All rights reserved.

package redact

// redactionSafe last 4 digits are usually considered safe (e.g. credit cards, iban, ...)
const redactionSafe = 4

var Default *PatternRedactor

func init() {
	scheme := RedactionSchemeKeepLastJWTNoSignature(redactionSafe)
	Default = NewPatternRedactor(scheme)
	Default.AddPatterns(AllPatterns...)
}
