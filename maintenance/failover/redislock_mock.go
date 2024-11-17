package failover

import (
	"context"
	"fmt"
	"github.com/bsm/redislock"
	"github.com/stretchr/testify/mock"
	"sync"
	"time"
)

// MockRedisLockClient represents a mock for redislock.Client
type MockRedisLockClient struct {
	mock.Mock
	failTime   time.Time
	failPeriod time.Duration
	mutex      sync.Mutex
}

func (m *MockRedisLockClient) Obtain(ctx context.Context, key string, ttl time.Duration, opt *redislock.Options) (*redislock.Lock, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Simulate failure during the failure window
	if time.Now().After(m.failTime) && time.Now().Before(m.failTime.Add(m.failPeriod)) {
		return nil, fmt.Errorf("network error during Obtain")
	}

	// Simulate successful lock acquisition
	mockLock := &MockLock{
		Lock:       &redislock.Lock{}, // Use an empty redislock.Lock
		ttl:        ttl,
		failTime:   m.failTime,
		failPeriod: m.failPeriod,
	}

	return mockLock.Lock, nil
}

// SetFailConditions sets the failure conditions for the mock client
func (m *MockRedisLockClient) SetFailConditions(failTime time.Time, failPeriod time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.failTime = failTime
	m.failPeriod = failPeriod
}
