// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/18 by Charlotte Pröller

package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
)

// HealthCheck checks the state of a redis connection. It must not be changed
// after it was registered as a health check.
type HealthCheck struct {
	state  servicehealthcheck.ConnectionState
	Client *redis.Client
}

// HealthCheck checks if the redis is healthy. If the last result is outdated,
// redis is checked for writeability and readability,
// otherwise return the old result
func (h *HealthCheck) HealthCheck(ctx context.Context) servicehealthcheck.HealthCheckResult {
	client := h.Client.WithContext(ctx)

	if time.Since(h.state.LastChecked()) <= cfg.HealthCheckResultTTL {
		// the last health check is not outdated, an can be reused.
		return h.state.GetState()
	}

	// Try writing
	if err := client.Set(cfg.HealthCheckKey, "true", 0).Err(); err != nil {
		h.state.SetErrorState(err)
		return h.state.GetState()
	}
	// If writing worked try reading
	err := client.Get(cfg.HealthCheckKey).Err()
	if err != nil {
		h.state.SetErrorState(err)
		return h.state.GetState()
	}
	// If reading an writing worked set the Health Check to healthy
	h.state.SetHealthy()
	return h.state.GetState()
}
