// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.

package redis

import (
	"context"
	"time"

	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
	"github.com/redis/go-redis/v9"
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
	if time.Since(h.state.LastChecked()) <= cfg.HealthCheckResultTTL {
		// the last health check is not outdated, an can be reused.
		return h.state.GetState()
	}

	// Try writing
	if err := h.Client.Set(ctx, cfg.HealthCheckKey, "true", 0).Err(); err != nil {
		h.state.SetErrorState(err)
		return h.state.GetState()
	}
	// If writing worked try reading
	err := h.Client.Get(ctx, cfg.HealthCheckKey).Err()
	if err != nil {
		h.state.SetErrorState(err)
		return h.state.GetState()
	}
	// If reading an writing worked set the Health Check to healthy
	h.state.SetHealthy()
	return h.state.GetState()
}
