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

const waitRetry = time.Millisecond * 500

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
	client *k8sapi.Client

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
func NewActivePassive(clusterName string, timeToFailover time.Duration, client *redis.Client) (*ActivePassive, error) {
	cl, err := k8sapi.NewClient()
	if err != nil {
		return nil, err
	}

	ap := &ActivePassive{
		clusterName:    clusterName,
		timeToFailover: timeToFailover,
		locker:         redislock.New(client),
		client:         cl,
	}
	health.SetCustomReadinessCheck(ap.Handler)

	return ap, nil
}

// Run registers the readiness probe and calls the OnActive
// and OnPassive callbacks in case the election took place.
// Will handle panic safely and therefore can be directly called
// with go.
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

	// Ticker that tries to refresh the lock's TTL well before the TTL actually expires to reduce the possibility of
	// small network delays or redis unavailability leading to a refresh try after the TTL has already expired.
	t := time.NewTicker(a.timeToFailover / 2)

	// Ticker to check if the lock can be acquired
	retry := time.NewTicker(waitRetry)

	for {
		// Allow close or cancel
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-a.close:
			return nil
		case <-t.C:
			if a.getState() == ACTIVE {
				err := lock.Refresh(ctx, a.timeToFailover, &redislock.Options{
					RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(a.timeToFailover/5), 2),
				})
				if err != nil {
					logger.Warn().Err(err).Msg("failed to refresh the redis lock; attempting to reacquire lock")

					// Attempt to reacquire the lock immediately but try only once
					var errReacquire error

					lock, errReacquire = a.locker.Obtain(ctx, lockName, a.timeToFailover, &redislock.Options{})
					if errReacquire == nil {
						// Successfully reacquired the lock, remain active
						logger.Info().Msg("redis lock reacquired after refresh failure; remaining active")

						a.becomeActive(ctx)
						t.Reset(a.timeToFailover / 2)
					} else {
						// We were active but couldn't refresh the lock TTL and reacquire the lock, so, become undefined
						logger.Info().Err(err).Msg("failed to reacquire the redis lock; becoming undefined")

						a.becomeUndefined(ctx)
					}
				}
			}
		case <-retry.C:
			// Try to acquire the lock as we are not active
			if a.getState() != ACTIVE {
				var err error
				lock, err = a.locker.Obtain(ctx, lockName, a.timeToFailover, &redislock.Options{
					RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(waitRetry/2), 1),
				})
				if err != nil {
					// couldn't obtain the lock; becoming passive
					if a.getState() != PASSIVE {
						logger.Info().Err(err).Msg("couldn't obtain the redis lock; becoming passive")
						a.becomePassive(ctx)
					}

					continue
				}

				// Lock acquired
				logger.Debug().Msg("redis lock acquired; becoming active")
				a.becomeActive(ctx)

				// We are active; renew the lock TTL if required
				d, err := lock.TTL(ctx)
				if err != nil {
					logger.Info().Err(err).Msg("failed to get TTL from redis lock")
				}
				if d == 0 {
					// TTL seems to be expired, retry to get lock or become passive in next iteration
					logger.Info().Msg("redis lock TTL has expired; becoming undefined")
					a.becomeUndefined(ctx)
				}
				refreshTime := d / 2

				logger.Debug().Msgf("redis lock TTL is still valid; set refresh time to %v ms", refreshTime)

				// Trigger a refresh after TTL / 2
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
	err := a.client.SetCurrentPodLabel(ctx, Label, a.label(state))
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
