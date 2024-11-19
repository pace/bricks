// Copyright © 2022 by PACE Telematics GmbH. All rights reserved.

package failover

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bsm/redislock"
	"github.com/pace/bricks/backend/k8sapi"
	"github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/health"
	"github.com/pace/bricks/maintenance/log"
	"github.com/redis/go-redis/v9"
)

type status int

const (
	PASSIVE   status = -1
	UNDEFINED status = 0
	ACTIVE    status = 1
)

const Label = "github.com.pace.bricks.activepassive"

// ActivePassive implements a fail over mechanism that allows
// to deploy a service multiple times but ony one will accept
// traffic by using the label selector of kubernetes.
// In order to determine the active, a lock needs to be hold
// in redis. Hooks can be passed to handle the case of becoming
// the active or passive.
// The readiness probe will report the state (ACTIVE/PASSIVE)
// of each of the members in the cluster.
type ActivePassive struct {
	// OnActive will be called in case the current process is elected to be the active one
	OnActive func()

	// OnPassive will be called in case the current process is the passive one
	OnPassive func()

	// OnStop is called after the ActivePassive process stops
	OnStop func()

	close          chan struct{}
	clusterName    string
	timeToFailover time.Duration
	locker         *redislock.Client

	// access to the kubernetes api
	k8sClient *k8sapi.Client

	// current status of the fail over (to show it in the readiness status)
	state   status
	stateMu sync.RWMutex
}

// NewActivePassive creates a new active passive cluster
// identified by the name. The time to fail over determines
// the frequency of checks performed against redis to
// keep the active state.
// NOTE: creating multiple ActivePassive in one process
// is not working correctly as there is only one readiness
// probe.
func NewActivePassive(clusterName string, timeToFailover time.Duration, redisClient *redis.Client) (*ActivePassive, error) {
	k8sClient, err := k8sapi.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the kubernets client: %w", err)
	}

	if redisClient == nil {
		return nil, fmt.Errorf("redis client is not initialized")
	}

	ap := &ActivePassive{
		clusterName:    clusterName,
		timeToFailover: timeToFailover,
		locker:         redislock.New(redisClient),
		k8sClient:      k8sClient,
	}

	health.SetCustomReadinessCheck(ap.Handler)

	return ap, nil
}

// Run manages distributed lock-based leadership.
// This method is designed to continually monitor and maintain the leadership status of the calling pod,
// ensuring only one active instance holds the lock at a time, while transitioning other instances to passive
// mode. The handler will try to renew its active status by refreshing the lock periodically, and attempt
// reacquisition on failure, avoiding potential race conditions for leadership.
func (a *ActivePassive) Run(ctx context.Context) error {
	defer errors.HandleWithCtx(ctx, "activepassive failover handler")

	lockName := "activepassive:lock:" + a.clusterName
	logger := log.Ctx(ctx).With().Str("failover", lockName).Logger()
	ctx = logger.WithContext(ctx)

	a.close = make(chan struct{})
	defer close(a.close)

	// Trigger stop handler
	defer func() {
		if a.OnStop != nil {
			a.OnStop()
		}
	}()

	var lock *redislock.Lock

	// Ticker to refresh the lock's TTL before it expires
	refreshInterval := a.timeToFailover / 2
	refresh := time.NewTicker(refreshInterval)
	logger.Debug().Msgf("Stefan: refresh interval: %v s", refreshInterval)

	// Ticker to check if the lock can be acquired if in passive or undefined state
	retryInterval := a.timeToFailover / 3
	retry := time.NewTicker(retryInterval)
	logger.Debug().Msgf("Stefan: retry interval: %v s", retryInterval)

	for {
		// Allow close or cancel
		select {
		case <-ctx.Done():
			logger.Warn().Err(ctx.Err()).Msg("Stefan: context canceled; exiting Run()")
			return ctx.Err()
		case <-a.close:
			logger.Warn().Err(ctx.Err()).Msg("Stefan: closed; exiting Run()")
			return nil
		case <-refresh.C:
			logger.Debug().Msgf("Stefan: tick from refresh; state: %v", a.getState())

			if a.getState() == ACTIVE && lock != nil {
				err := lock.Refresh(ctx, a.timeToFailover, &redislock.Options{
					RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(refreshInterval/5), 2),
				})
				if err != nil {
					logger.Warn().Err(err).Msg("Stefan: failed to refresh the redis lock; attempting to reacquire lock")

					// Attempt to reacquire the lock immediately with short and limited retries
					var errReacquire error

					lock, errReacquire = a.locker.Obtain(ctx, lockName, a.timeToFailover, &redislock.Options{
						RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(100*time.Millisecond), 2),
					})
					if errReacquire == nil {
						// Successfully reacquired the lock, remain active
						logger.Debug().Msg("Stefan: redis lock reacquired after refresh failure; remaining active")
						a.becomeActive(ctx)
						refresh.Reset(refreshInterval)
					} else {
						// We were active but couldn't refresh the lock TTL and reacquire the lock, so, become undefined
						logger.Debug().Err(err).Msg("Stefan: failed to reacquire the redis lock; becoming undefined")
						a.becomeUndefined(ctx)
					}
				} else {
					logger.Debug().Err(err).Msg("Stefan: redis lock refreshed")
				}
			}
		case <-retry.C:
			logger.Debug().Msgf("Stefan: tick from retry; state: %v", a.getState())

			// Try to acquire the lock as we are not active
			if a.getState() != ACTIVE {
				logger.Debug().Msgf("Stefan: in retry: trying to acquire the lock...")

				var err error
				lock, err = a.locker.Obtain(ctx, lockName, a.timeToFailover, &redislock.Options{
					RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(retryInterval/2), 2),
				})
				if err != nil {
					logger.Debug().Err(err).Msgf("Stefan: failed to obtain the redis lock; current state: %v", a.getState())

					// couldn't obtain the lock; becoming passive
					if a.getState() != PASSIVE {
						logger.Debug().Err(err).Msg("Stefan: couldn't obtain the redis lock; becoming passive")
						a.becomePassive(ctx)
					}

					continue
				}

				// Lock acquired, transitioning to active
				logger.Debug().Msg("Stefan: redis lock acquired; becoming active")
				a.becomeActive(ctx)

				logger.Debug().Msg("Stefan: became active")

				if lock == nil {
					logger.Debug().Msg("Stefan: lock is nil")
					a.becomeUndefined(ctx)
					continue
				}

				// Check TTL of the newly acquired lock and adjust refresh timer
				timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
				ttl, err := lock.TTL(timeoutCtx)
				if err != nil {
					// If trying to get the TTL from the lock fails we become undefined and retry acquisition at the next tick.
					logger.Debug().Err(err).Msg("Stefan: failed to get TTL from redis lock; becoming undefined")
					cancel()
					a.becomeUndefined(ctx)
					continue
				}

				cancel()

				logger.Debug().Msgf("Stefan: got TTL from redis lock: %v", ttl.Abs())

				if ttl == 0 {
					// Since the lock is very fresh with a TTL well > 0 this case is just a safeguard against rare occasions.
					logger.Debug().Msg("Stefan: redis lock TTL has expired; becoming undefined")
					a.becomeUndefined(ctx)
				} else {
					logger.Debug().Msg("Stefan: redis lock TTL > 0")

					// Enforce a minimum refresh time
					minRefreshTime := 2 * time.Second
					refreshTime := ttl / 2
					if refreshTime < minRefreshTime {
						logger.Warn().Msgf("Stefan: calculated refresh time %v is below minimum threshold; using %v instead", refreshTime, minRefreshTime)
						refreshTime = minRefreshTime
					}

					logger.Debug().Msgf("Stefan: redis lock TTL is still valid; set refresh time to %v ms", refreshTime)
					refresh.Reset(refreshTime)
				}
			}
		}
	}
}

// Stop stops acting as a passive or active member.
func (a *ActivePassive) Stop() {
	a.close <- struct{}{}
}

// Handler implements the readiness http endpoint
func (a *ActivePassive) Handler(w http.ResponseWriter, r *http.Request) {
	label := a.label(a.getState())
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, strings.ToUpper(label))
}

func (a *ActivePassive) label(s status) string {
	switch s {
	case ACTIVE:
		return "active"
	case PASSIVE:
		return "passive"
	default:
		return "undefined"
	}
}

func (a *ActivePassive) becomeActive(ctx context.Context) {
	if a.setState(ctx, ACTIVE) {
		if a.OnActive != nil {
			a.OnActive()
		}
	}
}

func (a *ActivePassive) becomePassive(ctx context.Context) {
	if a.setState(ctx, PASSIVE) {
		if a.OnPassive != nil {
			a.OnPassive()
		}
	}
}

func (a *ActivePassive) becomeUndefined(ctx context.Context) {
	a.setState(ctx, UNDEFINED)
}

// setState returns true if the state was set successfully
func (a *ActivePassive) setState(ctx context.Context, state status) bool {
	err := a.k8sClient.SetCurrentPodLabel(ctx, Label, a.label(state))
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to mark pod as undefined")
		a.stateMu.Lock()
		a.state = UNDEFINED
		a.stateMu.Unlock()
		return false
	}
	a.stateMu.Lock()
	a.state = state
	a.stateMu.Unlock()
	return true
}

func (a *ActivePassive) getState() status {
	a.stateMu.RLock()
	state := a.state
	a.stateMu.RUnlock()
	return state
}
