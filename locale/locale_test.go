// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package locale

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmpty(t *testing.T) {
	l := new(Locale)

	assert.False(t, l.HasLanguage())
	assert.False(t, l.HasTimezone())
	assert.Equal(t, None, l.Language())
	assert.Equal(t, None, l.Timezone())

	_, err := l.Location()
	assert.Error(t, err)
}

func TestLanguage(t *testing.T) {
	l := NewLocale("fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5", None)

	assert.True(t, l.HasLanguage())
	assert.False(t, l.HasTimezone())
	assert.Equal(t, "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5", l.Language())
	assert.Equal(t, None, l.Timezone())

	_, err := l.Location()
	assert.Error(t, err)
}

func TestTimezone(t *testing.T) {
	l := NewLocale(None, "Europe/Berlin")

	assert.False(t, l.HasLanguage())
	assert.True(t, l.HasTimezone())
	assert.Equal(t, None, l.Language())
	assert.Equal(t, "Europe/Berlin", l.Timezone())

	loc, err := l.Location()
	assert.NoError(t, err)
	timeInUTC := time.Date(2018, 8, 30, 12, 0, 0, 0, time.UTC)
	assert.Equal(t, "2018-08-30 14:00:00 +0200 CEST", timeInUTC.In(loc).String())
}

func TestTimezoneAndLocale(t *testing.T) {
	l := NewLocale("fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5", "Europe/Berlin")

	assert.True(t, l.HasLanguage())
	assert.True(t, l.HasTimezone())
	assert.Equal(t, "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5", l.Language())
	assert.Equal(t, "Europe/Berlin", l.Timezone())

	loc, err := l.Location()
	assert.NoError(t, err)
	timeInUTC := time.Date(2018, 8, 30, 12, 0, 0, 0, time.UTC)
	assert.Equal(t, "2018-08-30 14:00:00 +0200 CEST", timeInUTC.In(loc).String())
}

func TestLocale_Now(t *testing.T) {
	location, err := time.LoadLocation("America/Marigot")
	require.NoError(t, err)

	// Note: this test rounds to second
	tests := []struct {
		name           string
		acceptTimezone string
		want           string
	}{
		{"invalid locale", "Foo/bar", time.Now().Round(time.Minute).String()},
		{"valid locale", "America/Marigot", time.Now().Round(time.Minute).In(location).String()},
		{"no locale", None, time.Now().Round(time.Minute).String()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := Locale{acceptTimezone: tt.acceptTimezone}
			got := l.Now().Round(time.Minute)
			assert.Equal(t, tt.want, got.String())
		})
	}
}
