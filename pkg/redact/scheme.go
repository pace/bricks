// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/12/18 by Vincent Landgraf

package redact

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
	// WIP: implement removal of JWT signature, allows inspection but prevents stealing of jwt
	return RedactionSchemeKeepLast(num)
}
