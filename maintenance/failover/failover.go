// Copyright © 2022 by PACE Telematics GmbH. All rights reserved.

package failover

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
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

// ActivePassive implements a failover mechanism that allows
// to deploy a service multiple times but ony one will accept
// traffic by using the label selector of kubernetes.
// In order to determine the active, a lock needs to be hold
// in redis. Hooks can be passed to handle the case of becoming
// the active or passive.
// The readiness probe will report the state (ACTIVE/PASSIVE)
// of each of the members in the cluster.
type ActivePassive struct {
	// OnActive will be called in case the current processes is elected to be the active one
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

	// current status of the failover (to show it in the readiness status)
	state   status
	stateMu sync.RWMutex
}

// NewActivePassive creates a new active passive cluster
// identified by the name. The time to fail over determines
// the frequency of checks performed against redis to
// keep the active state.
// NOTE: creating multiple ActivePassive in one process
// is not working correctly as there is only one readiness probe.
func NewActivePassive(clusterName string, timeToFailover time.Duration, redisClient *redis.Client) (*ActivePassive, error) {
	k8sClient, err := k8sapi.NewClient()
	if err != nil {
		return nil, err
	}

	activePassive := &ActivePassive{
		clusterName:    clusterName,
		timeToFailover: timeToFailover,
		locker:         redislock.New(redisClient),
		k8sClient:      k8sClient,
	}
	health.SetCustomReadinessCheck(activePassive.Handler)

	return activePassive, nil
}

// Run manages distributed lock-based leadership.
// This method is designed to continually monitor and maintain the leadership status of the calling pod,
// ensuring only one active instance holds the lock at a time, while transitioning other instances to passive
// mode. The handler will try to renew its active status by refreshing the lock periodically.
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
	tryRefreshLock := time.NewTicker(a.timeToFailover)

	// Ticker to check if the lock can be acquired if in passive or undefined state
	retryInterval := 500 * time.Millisecond
	retryAcquireLock := time.NewTicker(retryInterval)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-a.close:
			return nil
		case <-tryRefreshLock.C:
			if a.getState() == ACTIVE {
				err := lock.Refresh(ctx, a.timeToFailover, &redislock.Options{
					RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(a.timeToFailover/3), 3),
				})
				if err != nil {
					logger.Info().Err(err).Msg("failed to refresh the lock; becoming undefined...")
					a.becomeUndefined(ctx)
				}
			}
		case <-retryAcquireLock.C:
			if a.getState() != ACTIVE {
				var err error

				lock, err = a.locker.Obtain(ctx, lockName, a.timeToFailover, &redislock.Options{
					RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(a.timeToFailover/3), 3),
				})
				if err != nil {
					// Couldn't obtain the lock; becoming passive
					if a.getState() != PASSIVE {
						logger.Debug().Err(err).Msg("becoming passive")
						a.becomePassive(ctx)
					}

					continue
				}

				// Lock acquired, transitioning to active
				logger.Debug().Msg("becoming active")
				a.becomeActive(ctx)

				// Check TTL of the newly acquired lock
				ttl, err := safeGetTTL(ctx, lock, logger)
				if err != nil {
					logger.Info().Err(err).Msg("failed to get activepassive lock TTL")
				}

				if ttl == 0 {
					// Since the lock is very fresh with a TTL well > 0 this case is just a safeguard against rare occasions.
					logger.Info().Msg("activepassive lock TTL is expired although the lock has been just acquired; becoming undefined...")
					a.becomeUndefined(ctx)
				}

				refreshTime := ttl / 2

				logger.Debug().Msgf("set refresh ticker to %v ms", refreshTime)

				// Reset the refresh ticker to TTL / 2
				tryRefreshLock.Reset(refreshTime)
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

// safeGetTTL tries to get the TTL from the provided redis lock and recovers from a panic inside TTL().
func safeGetTTL(ctx context.Context, lock *redislock.Lock, logger zerolog.Logger) (time.Duration, error) {
	var (
		err error
		ttl time.Duration
	)

	defer func() {
		if r := recover(); r != nil {
			logger.Error().Msgf("Recovered from panic in lock.TTL(): %v", r)
			err = fmt.Errorf("panic during lock.TTL(): %v", r)
		}
	}()

	ttl, err = lock.TTL(ctx)
	return ttl, err
}
