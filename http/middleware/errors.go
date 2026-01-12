// Copyright © 2020–2026 by PACE Mobility GmbH. All rights reserved.

package middleware

import "errors"

// All exported package errors.
var (
	ErrNotFound       = errors.New("not found")
	ErrInvalidRequest = errors.New("request is invalid")
)
