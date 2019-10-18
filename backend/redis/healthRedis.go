package redis

import (
	"github.com/go-redis/redis"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
	"sync"
	"time"
)

type redisHealthCheck struct {
	State  *servicehealthcheck.ConnectionState
	client *redis.Client
}

func (h *redisHealthCheck) InitHealthCheck() error {
	h.client = Client()
	h.State = &servicehealthcheck.ConnectionState{M: sync.Mutex{}}
	return nil
}

func (h *redisHealthCheck) Name() string {
	return "redis"
}

// HealthCheck if the last result is outdated, redis is checked for writeability and readability,
// otherwise return the old result
func (h *redisHealthCheck) HealthCheck(currTime time.Time) error {
	h.State.M.Lock()
	defer h.State.M.Unlock()
	if currTime.Sub(h.State.LastCheck) > cfg.HealthMaxRequest {
		//Set initial to healthy:
		h.State.IsHealthy = true
		h.State.Err = nil
		h.State.LastCheck = currTime
		//Try writing:
		errWrite := h.client.Append(cfg.HealthKey, "true").Err()
		if errWrite != nil {
			h.State.IsHealthy = false
			h.State.Err = errWrite

		} else {
			//If writing worked try reading
			errRead := h.client.Get(cfg.HealthKey).Err()
			h.State.IsHealthy = errRead == nil
			h.State.Err = errRead
		}

	}
	return h.State.Err
}

func (h *redisHealthCheck) CleanUp() error {
	//Nop, nothing to cleanup
	return nil

}
