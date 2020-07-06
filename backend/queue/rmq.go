package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/adjust/rmq/v2"
	"github.com/pace/bricks/backend/redis"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
)

var (
	rmqConnection     rmq.Connection
	queueHealthLimits sync.Map

	initMutex sync.Mutex
)

func initDefault() {
	initMutex.Lock()
	defer initMutex.Unlock()

	if rmqConnection != nil {
		return
	}

	rmqConnection = rmq.OpenConnectionWithRedisClient("default", redis.Client())
	gatherMetrics(rmqConnection)
	servicehealthcheck.RegisterHealthCheck("rmq", &HealthCheck{})
}

// NewQueue creates a new rmq.Queue and initializes health checks for this queue
// Whenever the number of items in the queue exceeds the healthyLimit
// The queue will be reported as unhealthy
// If the queue has already been opened, it will just be returned. Limits will not
// be updated
func NewQueue(name string, healthyLimit int) rmq.Queue {
	initDefault()
	queue := rmqConnection.OpenQueue(name)
	if _, ok := queueHealthLimits.Load(name); ok {
		return queue
	}
	queueHealthLimits.Store(name, healthyLimit)
	return queue
}

type HealthCheck struct {
	state servicehealthcheck.ConnectionState
	// IgnoreInterval is a switch used for testing, just to allow multiple
	// functional queries of HealthCheck in rapid bursts
	IgnoreInterval bool
}

// HealthCheck checks if the queues are healthy, i.e. whether the number of
// items accumulated is below the healthyLimit defined when opening the queue
func (h *HealthCheck) HealthCheck(ctx context.Context) servicehealthcheck.HealthCheckResult {
	if !h.IgnoreInterval && time.Since(h.state.LastChecked()) <= cfg.HealthCheckResultTTL {
		return h.state.GetState()
	}

	stats := rmqConnection.CollectStats(rmqConnection.GetOpenQueues())
	queueHealthLimits.Range(func(k, v interface{}) bool {
		name := k.(string)
		healthLimit := v.(int)
		stat := stats.QueueStats[name]
		if stat.ReadyCount > healthLimit {
			h.state.SetErrorState(fmt.Errorf("Queue '%s' exceeded safe health limit of '%d'", name, healthLimit))
			return false
		}
		h.state.SetHealthy()
		return true
	})
	return h.state.GetState()
}
