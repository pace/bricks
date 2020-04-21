// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/04/16 by Vincent Landgraf

package locale

import (
	"container/list"
	"context"
)

// Strategy defines a function that returns a Locale based on the passed Context
type Strategy func(ctx context.Context) *Locale

// NewContextStrategy returns a strategy that defines a static fallback language and timezone.
// If only lang or timezone fallback should be defined as a fallback, the None value may be used.
func NewFallbackStrategy(lang, timezone string) Strategy {
	l := NewLocale(lang, timezone)
	return func(ctx context.Context) *Locale {
		return l
	}
}

// NewContextStrategy returns a strategy that takes the locale form the request
func NewContextStrategy() Strategy {
	return func(ctx context.Context) *Locale {
		l, _ := FromCtx(ctx)
		return l
	}
}

// StrategyList has a list of strategies that are evaluated to find
// the correct user locale
type StrategyList struct {
	strategies list.List
}

// PushBack inserts the passed strategies at the back of list
func (s *StrategyList) PushBack(strategies ...Strategy) {
	for _, strategy := range strategies {
		s.strategies.PushBack(strategy)
	}
}

// PushFront inserts a passed strategies at the front of list
func (s *StrategyList) PushFront(strategies ...Strategy) {
	for _, strategy := range strategies {
		s.strategies.PushFront(strategy)
	}
}

// Locale executes all strategies and returns the new locale
func (s *StrategyList) Locale(ctx context.Context) *Locale {
	var l Locale
	for i := s.strategies.Front(); i != nil; i = i.Next() {
		curLoc := (i.Value.(Strategy))(ctx)

		// take language if defined
		if !l.HasLanguage() && curLoc.HasLanguage() {
			l.acceptLanguage = curLoc.Language()
		}

		// take timezone if defined
		if !l.HasTimezone() && curLoc.HasTimezone() {
			l.acceptTimezone = curLoc.Timezone()
		}

		// if locale is complete stop chain
		if l.HasLanguage() && l.HasTimezone() {
			break
		}
	}
	return &l
}

// NewDefaultFallbackStrategy returns a strategy list configured via environment
func NewDefaultFallbackStrategy() *StrategyList {
	var sl StrategyList
	sl.PushFront(NewFallbackStrategy(cfg.Language, cfg.Timezone))
	sl.PushFront(NewContextStrategy())
	return &sl
}
