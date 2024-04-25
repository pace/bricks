// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/08/12 by Marius Neugebauer

package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var _ Cache = (*Redis)(nil)

// Redis is the cache that uses a redis backend. It is safe for concurrent use.
type Redis struct {
	client *redis.Client
	prefix string
}

// InRedis returns a new cache that connects to redis using the given client.
// The prefix is used for every key that is stored.
func InRedis(client *redis.Client, prefix string) *Redis {
	return &Redis{
		client: client,
		prefix: prefix,
	}
}

// Put stores the value under the key. Any existing value is overwritten. If ttl
// is given, the cache automatically forgets the value after the duration. If
// ttl is zero then it is never automatically forgotten.
func (c *Redis) Put(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	err := c.client.Set(ctx, c.prefix+key, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("%w: redis: %s", ErrBackend, err)
	}
	return nil
}

// Lua script for Redis that returns both the value and the TTL in milliseconds
// of any key.
var redisGETAndPTTL = redis.NewScript(`return {
	redis.call('get',  KEYS[1]),
	redis.call('pttl', KEYS[1]),
}`)

// Get returns the value stored under the key and its remaining ttl. If there is
// no value stored, ErrNotFound is returned. If the ttl is zero, the value does
// not automatically expire. Unless an error is returned, the value is always
// non-nil.
func (c *Redis) Get(ctx context.Context, key string) ([]byte, time.Duration, error) {
	key = c.prefix + key
	r, err := redisGETAndPTTL.Run(ctx, c.client, []string{key}).Result()
	if err != nil {
		return nil, 0, fmt.Errorf("%w: redis: %s", ErrBackend, err)
	}
	result, ok := r.([]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("%w: redis returned unexpected type %T, expected %T", ErrBackend, r, result)
	}
	v := result[0]
	if v == nil {
		return nil, 0, fmt.Errorf("key %q: %w", key, ErrNotFound)
	}
	value, ok := v.(string)
	if !ok {
		return nil, 0, fmt.Errorf("%w: redis returned unexpected type %T, expected %T", ErrBackend, v, value)
	}
	ttl, ok := result[1].(int64)
	if !ok {
		return nil, 0, fmt.Errorf("%w: redis returned unexpected type %T, expected %T", ErrBackend, result[1], ttl)
	}
	switch {
	case ttl == -1: // key exists but has no associated expire
		return []byte(value), 0, nil
	case ttl == 0: // about to expire this millisecond
		return []byte(value), time.Duration(1), nil // use smallest non-zero duration
	case ttl > 0: // ttl is in ms
		return []byte(value), time.Duration(ttl) * time.Millisecond, nil
	default: // some error
		return nil, 0, fmt.Errorf("%w: redis: pttl returned %d", ErrBackend, ttl)
	}
}

// Forget removes the value stored under the key. No error is returned if there
// is no value stored.
func (c *Redis) Forget(ctx context.Context, key string) error {
	err := c.client.Del(ctx, c.prefix+key).Err()
	if err != nil {
		return fmt.Errorf("%w: redis: %s", ErrBackend, err)
	}
	return nil
}
