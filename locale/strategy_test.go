// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/04/16 by Vincent Landgraf

package locale

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStrategy(t *testing.T) {
	sl := NewDefaultFallbackStrategy()
	s := sl.Locale(context.Background())
	assert.Equal(t, "de-DE", s.Language())
	assert.Equal(t, "Europe/Berlin", s.Timezone())
}

func TestStrategyWithCtx(t *testing.T) {
	var sl StrategyList
	sl.PushBack(
		NewContextStrategy(),
		NewFallbackStrategy("de-DE", "Europe/Berlin"),
	)

	s := sl.Locale(WithLocale(context.Background(), NewLocale("fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5", "Europe/Paris")))
	assert.Equal(t, "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5", s.Language())
	assert.Equal(t, "Europe/Paris", s.Timezone())
}
