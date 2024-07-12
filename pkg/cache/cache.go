// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package cache

import (
	"context"
	"time"
)

// Cache is a common abstraction to cache bytes. It is safe for concurrent use.
type Cache interface {
	// Put stores the value under the key. Any existing value is overwritten. If
	// ttl is given, the cache automatically forgets the value after the
	// duration. If ttl is zero then it is never automatically forgotten.
	Put(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Get returns the value stored under the key and its remaining ttl. If
	// there is no value stored, ErrNotFound is returned. If the ttl is zero,
	// the value does not automatically expire. Unless an error is returned, the
	// value is always non-nil.
	Get(ctx context.Context, key string) (value []byte, ttl time.Duration, _ error)

	// Forget removes the value stored under the key. No error is returned if
	// there is no value stored.
	Forget(ctx context.Context, key string) error
}
