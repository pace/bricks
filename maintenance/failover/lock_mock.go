package failover

import (
	"context"
	"fmt"
	"github.com/bsm/redislock"
	"time"
)

// MockLock represents a mock for redislock.Lock
type MockLock struct {
	*redislock.Lock // Embed redislock.Lock to satisfy the return type
	ttl             time.Duration
	failTime        time.Time
	failPeriod      time.Duration
}

func (m *MockLock) Refresh(ctx context.Context, ttl time.Duration, options *redislock.Options) error {
	if time.Now().After(m.failTime) && time.Now().Before(m.failTime.Add(m.failPeriod)) {
		return fmt.Errorf("network error during Refresh")
	}
	m.ttl = ttl
	return nil
}

func (m *MockLock) TTL(ctx context.Context) (time.Duration, error) {
	if time.Now().After(m.failTime) && time.Now().Before(m.failTime.Add(m.failPeriod)) {
		return 0, fmt.Errorf("network error during TTL")
	}
	return m.ttl, nil
}
