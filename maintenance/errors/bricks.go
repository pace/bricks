/*
 * Copyright © 2023 by PACE Telematics GmbH. All rights reserved.
 * Created at 2023/1/20 by Sascha Voth
 */

package errors

import (
	"strconv"

	uuid "github.com/satori/go.uuid"

	"github.com/pace/bricks/http/jsonapi/runtime"
)

// BricksError - a bricks err is a bricks specific error which provides
// convenience functions to be transformed into runtime.Errors (JSON errors)
// pb generate can be used to create a set of pre defined BricksErrors based
// on a JSON specification, see pb generate for details
type BricksError struct {
	// title - a short, human-readable summary of the problem that SHOULD NOT change from occurrence
	// to occurrence of the problem, except for purposes of localization.
	title string
	// detail - an application-specific error code, expressed as a string value.
	detail string
	// status - // the HTTP status code applicable to this problem, expressed as
	// an int value. This SHOULD be provided.
	status int
	// code - application specific error code
	code string
}

func NewBricksError(opts ...BricksErrorOption) *BricksError {
	e := &BricksError{}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *BricksError) Error() string {
	return e.code
}

func (e *BricksError) Status() int {
	return e.status
}

// AsRuntimeError - returns the BricksError as bricks runtime.Error which aligns
// with a JSON error object
func (e *BricksError) AsRuntimeError() *runtime.Error {
	j := &runtime.Error{
		ID:     uuid.NewV4().String(),
		Status: strconv.Itoa(e.status),
		Code:   e.code,
		Title:  e.title,
		Detail: e.detail,
	}
	return j
}

// Equals - verifies if the given runtime.Error code equals the BricksError code
func (e *BricksError) Equals(re *runtime.Error) bool {
	return re.Code == e.code
}

type BricksErrorOption func(e *BricksError)

func WithDetail(s string) BricksErrorOption {
	return func(e *BricksError) {
		e.detail = s
	}
}

func WithTitle(s string) BricksErrorOption {
	return func(e *BricksError) {
		e.title = s
	}
}

func WithCode(s string) BricksErrorOption {
	return func(e *BricksError) {
		e.code = s
	}
}

func WithStatus(s int) BricksErrorOption {
	return func(e *BricksError) {
		e.status = s
	}
}
