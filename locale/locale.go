// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

// The http locale package helps to transport and use the localization
// information in a microservice landscape. It enables the propagation
// of the locale information using contexts.
//
// Two important aspects of localization are part of this package
// the language (RFC 7231, section 5.3.5: Accept-Language) and Timezone
// (RFC 7808).
//
// In order to get the Timezone information the package defines a new
// HTTP header "Accept-Timezone" to present times with the requested
// preference
package locale

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// None is no timezone or language
const None = ""

var ErrNoTimezone = errors.New("no timezone given")

// Locale contains the preferred language and timezone of the request
type Locale struct {
	// Language as per RFC 7231, section 5.3.5: Accept-Language
	// Example: "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5"
	acceptLanguage string
	// Timezone as per RFC 7808
	// Example: "Europe/Paris"
	acceptTimezone string
}

// NewLocale creates a new locale based on the passed accepted and language
// and timezone
func NewLocale(acceptLanguage, acceptTimezone string) *Locale {
	return &Locale{
		acceptLanguage: acceptLanguage,
		acceptTimezone: acceptTimezone,
	}
}

// Language of the locale
func (l Locale) Language() string {
	return l.acceptLanguage
}

// HasTimezone returns true if the language is defined, false otherwise
func (l Locale) HasLanguage() bool {
	return l.acceptLanguage != None
}

// Timezone of the locale
func (l Locale) Timezone() string {
	return l.acceptTimezone
}

// HasTimezone returns true if the timezone is defined, false otherwise
func (l Locale) HasTimezone() bool {
	return l.acceptTimezone != None
}

// Location based of the locale timezone
func (l Locale) Location() (*time.Location, error) {
	if !l.HasTimezone() {
		return nil, ErrNoTimezone
	}

	return time.LoadLocation(l.Timezone())
}

const serializeSep = "|"

// Serialize into a transportable form
func (l Locale) Serialize() string {
	return l.acceptLanguage + serializeSep + l.acceptTimezone
}

// Now returns the current time with the set timezone
// or local time if timezone is not set
func (l Locale) Now() time.Time {
	if l.HasTimezone() {
		loc, err := l.Location()
		if err != nil { // if the tz doesn't exist
			return time.Now()
		}
		return time.Now().In(loc)
	}

	return time.Now() // Local
}

// ParseLocale parses a serialized locale
func ParseLocale(serialized string) (*Locale, error) {
	parts := strings.Split(serialized, serializeSep)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid locale format: %q", serialized)
	}
	return NewLocale(parts[0], parts[1]), nil
}
