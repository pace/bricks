package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/adjust/rmq/v3"
	"github.com/pace/bricks/backend/redis"
	pberrors "github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
	"github.com/pace/bricks/maintenance/log"
	"github.com/pace/bricks/pkg/routine"
)

var (
	rmqConnection     rmq.Connection
	queueHealthLimits sync.Map

	initMutex sync.Mutex
)

func initDefault() error {
	var err error
	initMutex.Lock()
	defer initMutex.Unlock()

	if rmqConnection != nil {
		return nil
	}

	errChan := make(chan error)

	ctx := log.ContextWithSink(log.WithContext(context.Background()), new(log.Sink))
	routine.Run(ctx, func(ctx context.Context) {
		for {
			err := <-errChan
			if err != nil {
				pberrors.Handle(ctx, fmt.Errorf("rmq reported error in background task"))
			}
		}
	})

	rmqConnection, err = rmq.OpenConnectionWithRedisClient("default", redis.Client(), errChan)
	if err != nil {
		rmqConnection = nil
		return err
	}
	gatherMetrics(rmqConnection)
	servicehealthcheck.RegisterHealthCheck("rmq", &HealthCheck{})
	return nil
}

// NewQueue creates a new rmq.Queue and initializes health checks for this queue
// Whenever the number of items in the queue exceeds the healthyLimit
// The queue will be reported as unhealthy
// If the queue has already been opened, it will just be returned. Limits will not
// be updated
func NewQueue(name string, healthyLimit int) (rmq.Queue, error) {
	err := initDefault()
	if err != nil {
		return nil, err
	}
	queue, err := rmqConnection.OpenQueue(name)
	if _, ok := queueHealthLimits.Load(name); ok {
		return queue, nil
	}
	queueHealthLimits.Store(name, healthyLimit)
	return queue, nil
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

	queues, err := rmqConnection.GetOpenQueues()
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("rmq HealthCheck: could not get open queues")
		h.state.SetErrorState(fmt.Errorf("error while retrieving open queues: %s", err))
		return h.state.GetState()
	}
	stats, err := rmqConnection.CollectStats(queues)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("rmq HealthCheck: could not collect stats")
		h.state.SetErrorState(fmt.Errorf("error while collecting stats: %s", err))
		return h.state.GetState()
	}
	queueHealthLimits.Range(func(k, v interface{}) bool {
		name := k.(string)
		healthLimit := v.(int)
		stat := stats.QueueStats[name]
		if stat.ReadyCount > int64(healthLimit) {
			h.state.SetErrorState(fmt.Errorf("Queue '%s' exceeded safe health limit of '%d'", name, healthLimit))
			return false
		}
		h.state.SetHealthy()
		return true
	})
	return h.state.GetState()
}
