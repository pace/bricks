package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/adjust/rmq/v2"
	"github.com/pace/bricks/backend/redis"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
)

var (
	rmqConnection     rmq.Connection
	queueHealthLimits map[string]int
)

func init() {
	rmqConnection = rmq.OpenConnectionWithRedisClient("default", redis.Client())
	gatherMetrics(rmqConnection)
	servicehealthcheck.RegisterHealthCheck("rmq", &HealthCheck{})

	queueHealthLimits = map[string]int{}
}

// NewQueue creates a new rmq.Queue and initializes health checks for this queue
// Whenever the number of items in the queue exceeds the healthyLimit
// The queue will be reported as unhealthy
// If the queue has already been opened, it will just be returned. Limits will not
// be updated
func NewQueue(name string, healthyLimit int) rmq.Queue {
	queue := rmqConnection.OpenQueue(name)
	if _, ok := queueHealthLimits[name]; ok {
		return queue
	}
	queueHealthLimits[name] = healthyLimit
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
	for k, healthLimit := range queueHealthLimits {
		stat := stats.QueueStats[k]
		if stat.ReadyCount > healthLimit {
			h.state.SetErrorState(fmt.Errorf("Queue '%s' exceeded safe health limit of '%d'", k, healthLimit))
			return h.state.GetState()
		}
	}
	h.state.SetHealthy()
	return h.state.GetState()
}
