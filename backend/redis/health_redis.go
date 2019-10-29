// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/18 by Charlotte Pröller

package redis

import (
	"time"

	"github.com/go-redis/redis"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
)

type redisHealthCheck struct {
	State  *servicehealthcheck.ConnectionState
	client *redis.Client
}

func (h *redisHealthCheck) InitHealthCheck() error {
	h.client = Client()
	h.State = servicehealthcheck.NewConnectionState()
	return nil
}

// HealthCheck checks if the redis is healthy. If the last result is outdated,
// redis is checked for writeability and readability,
// otherwise return the old result
func (h *redisHealthCheck) HealthCheck() (bool, error) {
	currTime := time.Now()

	if currTime.Sub(h.State.LastCheck) <= cfg.HealthMaxRequest {
		// the last health check is not outdated, an can be reused.
		return h.State.GetState()
	}
	// Try writing
	errWrite := h.client.Append(cfg.HealthKey, "true").Err()
	if errWrite != nil {
		h.State.SetErrorState(errWrite, currTime)
		return h.State.GetState()
	}
	// If writing worked try reading
	errRead := h.client.Get(cfg.HealthKey).Err()
	if errRead != nil {
		h.State.SetErrorState(errRead, currTime)
		return h.State.GetState()
	}
	// If reading an writing worked set the Health Check to healthy
	h.State.SetHealthy(currTime)
	return h.State.GetState()
}

func (h *redisHealthCheck) CleanUp() error {
	//Nop, nothing to cleanup
	return nil

}
