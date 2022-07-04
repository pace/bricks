// Copyright © 2021 by PACE Telematics GmbH. All rights reserved.
// Created at 2021/02/04 by Julius Foitzik

package runtime

import (
	"errors"
	"fmt"
	"strconv"

	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"

	"github.com/pace/bricks/pkg/isotime"
)

var _ ValueSanitizer = (*datetimeSanitizer)(nil)
var _ ValueSanitizer = (*intSanitizer)(nil)
var _ ValueSanitizer = (*decimalSanitizer)(nil)
var _ ValueSanitizer = (*noopSanitizer)(nil)
var _ ValueSanitizer = (*uuidSanitizer)(nil)
var _ ValueSanitizer = (*composableAndFieldRestrictedSanitizer)(nil)

var (
	ErrInvalidFieldname = errors.New("invalid fieldName, not registered in sanitizer")
)

type datetimeSanitizer struct{}

func (d datetimeSanitizer) SanitizeValue(fieldName string, value string) (interface{}, error) {
	t, err := isotime.ParseISO8601(value)
	if err != nil {
		return nil, err
	}
	return t, nil
}

type intSanitizer struct{}

func (i intSanitizer) SanitizeValue(fieldName string, value string) (interface{}, error) {
	return strconv.Atoi(value)
}

type decimalSanitizer struct{}

func (d decimalSanitizer) SanitizeValue(fieldName string, value string) (interface{}, error) {
	return decimal.NewFromString(value)
}

type noopSanitizer struct{}

func (n noopSanitizer) SanitizeValue(fieldName string, value string) (interface{}, error) {
	return value, nil
}

type uuidSanitizer struct{}

func (u uuidSanitizer) SanitizeValue(fieldName string, value string) (interface{}, error) {
	if _, err := uuid.FromString(value); err != nil {
		return nil, err
	}
	return value, nil
}

type composableAndFieldRestrictedSanitizer map[string]ValueSanitizer

func (c composableAndFieldRestrictedSanitizer) SanitizeValue(fieldName string, value string) (interface{}, error) {
	san, found := c[fieldName]
	if !found {
		return nil, fmt.Errorf("%w: %v", ErrInvalidFieldname, fieldName)
	}
	return san.SanitizeValue(fieldName, value)
}

func NewComposableSanitizer(mapping map[string]ValueSanitizer) ValueSanitizer {
	return composableAndFieldRestrictedSanitizer(mapping)
}

func NewDatetimeSanitizer() ValueSanitizer {
	return &datetimeSanitizer{}
}

func NewIntSanitizer() ValueSanitizer {
	return &intSanitizer{}
}
func NewNoopSanitizer() ValueSanitizer {
	return &noopSanitizer{}
}
func NewUUIDSanitizer() ValueSanitizer {
	return &uuidSanitizer{}
}
func NewDecimalSanitizer() ValueSanitizer {
	return &decimalSanitizer{}
}
