// Copyright © 2021–2026 by PACE Mobility GmbH. All rights reserved.

package redact

// redactionSafe last 4 digits are usually concidered safe (e.g. credit cards, iban, ...)
const redactionSafe = 4

var Default *PatternRedactor

func init() {
	scheme := RedactionSchemeKeepLastJWTNoSignature(redactionSafe)
	Default = NewPatternRedactor(scheme)
	Default.AddPatterns(AllPatterns...)
}
