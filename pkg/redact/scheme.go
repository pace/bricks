// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package redact

import "strings"

type RedactionScheme func(string) string

// RedactionSchemeDoNothing doesn't redact any values
// Note: only use for testing
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

// RedactionSchemeKeepLast replaces all runes in the string with an asterisk
// except the last NUM runes
func RedactionSchemeKeepLastJWTNoSignature(num int) func(string) string {
	defaultScheme := RedactionSchemeKeepLast(num)

	return func(s string) string {
		if PatternJWT.Match([]byte(s)) {
			parts := strings.Split(s, ".")
			parts[2] = defaultScheme(parts[2])
			return strings.Join(parts, ".")
		}

		return defaultScheme(s)
	}
}
