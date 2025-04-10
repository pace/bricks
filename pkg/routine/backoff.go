// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package routine

import (
	"time"

	exponential "github.com/jpillora/backoff"
)

// Manages several backoffs of which at any time only one or none is used. When
// getting the duration of one backoff, all others are reset.
type combinedExponentialBackoff map[any]*exponential.Backoff

// ResetAll resets all backoffs.
func (all combinedExponentialBackoff) ResetAll() {
	for _, backoff := range all {
		backoff.Reset()
	}
}

// Duration returns the duration of the requested backoff and resets all others.
func (all combinedExponentialBackoff) Duration(key any) (dur time.Duration) {
	for k, backoff := range all {
		if k == key {
			dur = backoff.Duration()
		} else {
			backoff.Reset()
		}
	}

	return
}
