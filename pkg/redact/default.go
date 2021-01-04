// Copyright © 2021 by PACE Telematics GmbH. All rights reserved.
// Created at 2021/01/04 by Vincent Landgraf

package redact

// redactionSafe last 4 digits are usually concidered safe (e.g. credit cards, iban, ...)
const redactionSafe = 4

var Default *PatternRedactor

func init() {
	scheme := RedactionSchemeKeepLastJWTNoSignature(redactionSafe)
	Default = NewPatternRedactor(scheme)
	Default.AddPatterns(AllPatterns...)
}
