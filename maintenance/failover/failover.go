// Copyright © 2022 by PACE Telematics GmbH. All rights reserved.
// Created at 2022/01/20 by Vincent Landgraf

package failover

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v7"
	"github.com/pace/bricks/maintenance/health"
	"github.com/pace/bricks/maintenance/log"
)

const waitRetry = time.Millisecond * 500

type status int

const (
	PASSIVE   status = -1
	UNDEFINED status = 0
	ACTIVE    status = 1
)

// ActivePassive implements a failover mechanism that allows
// to deploy a service multiple times but ony one will accept
// traffic by using the readiness check of kubernetes.
// In order to determine the active, a lock needs to be hold
// in redis. Hocks can be passed to handle the case of becoming
// the active or passive.
// The readiness probe will report the state (ACTIVE/PASSIVE)
// of each of the members in the cluster.
type ActivePassive struct {
	// OnActive will be called in case the current processes
	// is elected to be the active one
	OnActive func()

	// OnPassive will be called in case the current process is
	// the passive one
	OnPassive func()

	// OnStop is called after the ActivePassive process stops
	OnStop func()

	close          chan struct{}
	clusterName    string
	timeToFailover time.Duration
	locker         *redislock.Client

	state   status
	stateMu sync.RWMutex
}

// NewActivePassive creates a new active passive cluster
// identified by the name, the time to failover determines
// the frequency of checks performed against the redis to
// keep the active state.
// NOTE: creating multiple ActivePassive in one processes
// is not working correctly as there is only one readiness
// probe.
func NewActivePassive(clusterName string, timeToFailover time.Duration, client *redis.Client) *ActivePassive {
	ap := &ActivePassive{
		clusterName:    clusterName,
		timeToFailover: timeToFailover,
		locker:         redislock.New(client),
	}
	health.SetCustomReadinessCheck(ap.Handler)
	return ap
}

// Run registers the readiness probe and calls the OnActive
// and OnPassive callbacks in case the election toke place.
func (a *ActivePassive) Run(ctx context.Context) error {
	lockName := "activepassive:lock:" + a.clusterName
	logger := log.Ctx(ctx).With().Str("failover", lockName).Logger()

	a.close = make(chan struct{})
	defer close(a.close)

	// trigger stop handler
	defer func() {
		if a.OnStop != nil {
			a.OnStop()
		}
	}()

	var lock *redislock.Lock

	// t is a ticker that reminds to call refresh if
	// the token was acquired after half of the remaining ttl time
	t := time.NewTicker(a.timeToFailover)

	// retry time triggers to check if the look needs to be acquired
	retry := time.NewTicker(waitRetry)

	for {
		// allow close or cancel
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-a.close:
			return nil
		case <-t.C:
			if a.getState() == ACTIVE {
				err := lock.Refresh(a.timeToFailover, &redislock.Options{
					RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(a.timeToFailover/3), 3),
				})
				if err != nil {
					logger.Debug().Err(err).Msgf("failed to refresh")
					a.setState(UNDEFINED)
				}
			}
		case <-retry.C:
			// try to acquire the lock, as we are not the active
			if a.getState() != ACTIVE {
				var err error
				lock, err = a.locker.Obtain(lockName, a.timeToFailover, &redislock.Options{
					RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(a.timeToFailover/3), 3),
				})
				if err != nil {
					// we became passive, trigger callback
					if a.getState() != PASSIVE {
						logger.Debug().Err(err).Msg("becoming passive")
						a.setState(PASSIVE)
						if a.OnPassive != nil {
							a.OnPassive()
						}
					}

					continue
				}

				// lock acquired
				logger.Debug().Msg("becoming active")
				a.setState(ACTIVE)
				if a.OnActive != nil {
					a.OnActive()
				}

				// we are active, renew if required
				d, err := lock.TTL()
				if err != nil {
					logger.Debug().Err(err).Msgf("failed to get TTL %q")
				}
				if d == 0 {
					// TTL seems to be expired, retry to get lock or become
					// passive in next iteration
					a.setState(UNDEFINED)
					logger.Debug().Msg("ttl expired")
				}
				refreshTime := d / 2

				logger.Debug().Msgf("set refresh to %v", refreshTime)

				// set to trigger refresh after TTL / 2
				t.Reset(refreshTime)
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
	var str string
	var code int

	switch a.getState() {
	case UNDEFINED:
		str = "UNDEFINED"
		code = 503
	case ACTIVE:
		str = "ACTIVE"
		code = 200
	case PASSIVE:
		str = "PASSIVE"
		code = 502
	}

	w.WriteHeader(code)
	fmt.Fprintln(w, str)
}

func (a *ActivePassive) setState(state status) {
	a.stateMu.Lock()
	a.state = state
	a.stateMu.Unlock()
}

func (a *ActivePassive) getState() status {
	a.stateMu.RLock()
	state := a.state
	a.stateMu.RUnlock()
	return state
}
