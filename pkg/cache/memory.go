// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/08/12 by Marius Neugebauer

package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

var _ Cache = (*Memory)(nil)

// Memory is the cache that stores everything in memory.  It is safe for
// concurrent use.
type Memory struct {
	values map[string]inMemoryValue
	mx     sync.RWMutex
}

type inMemoryValue struct {
	value     []byte
	expiresAt time.Time
}

// InMemory returns a new in-memory cache.
func InMemory() *Memory {
	return &Memory{
		values: make(map[string]inMemoryValue, 1),
	}
}

// Put stores the value under the key. Any existing value is overwritten. If ttl
// is given, the cache automatically forgets the value after the duration. If
// ttl is zero then it is never automatically forgotten.
func (c *Memory) Put(_ context.Context, key string, value []byte, ttl time.Duration) error {
	v := inMemoryValue{value: make([]byte, len(value))}
	copy(v.value, value)
	if ttl != 0 {
		v.expiresAt = time.Now().Add(ttl)
	}
	c.mx.Lock()
	c.values[key] = v
	c.mx.Unlock()
	return nil
}

// Get returns the value stored under the key and its remaining ttl. If there is
// no value stored, ErrNotFound is returned. If the ttl is zero, the value does
// not automatically expire. Unless an error is returned, the value is always
// non-nil.
func (c *Memory) Get(ctx context.Context, key string) ([]byte, time.Duration, error) {
	c.mx.RLock()
	v, ok := c.values[key]
	c.mx.RUnlock()
	if !ok {
		return nil, 0, fmt.Errorf("key %q: %w", key, ErrNotFound)
	}
	var ttl time.Duration
	if !v.expiresAt.IsZero() {
		ttl = time.Until(v.expiresAt)
		if ttl <= 0 {
			c.forget(key)
			return nil, 0, fmt.Errorf("key %q: %w", key, ErrNotFound)
		}
	}
	value := make([]byte, len(v.value))
	copy(value, v.value)
	return value, ttl, nil
}

// Forget removes the value stored under the key. No error is returned if there
// is no value stored.
func (c *Memory) Forget(_ context.Context, key string) error {
	c.forget(key)
	return nil
}

func (c *Memory) forget(key string) {
	c.mx.Lock()
	delete(c.values, key)
	c.mx.Unlock()
}
