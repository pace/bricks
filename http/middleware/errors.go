// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/08/27 by Marius Neugebauer

package middleware

import "errors"

// All exported package errors.
var (
	ErrNotFound       = errors.New("not found")
	ErrInvalidRequest = errors.New("request is invalid")
)
