// Copyright © 2020–2026 by PACE Mobility GmbH. All rights reserved.

package cache

import "errors"

// Package errors.
var (
	// The value under the given key was not found.
	ErrNotFound = errors.New("not found")

	// The caching backend produced an error that is not reflected by any other
	// error.
	ErrBackend = errors.New("cache backend error")
)
