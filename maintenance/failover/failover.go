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
	OnActive func(ctx context.Context)

	// OnPassive will be called in case the current process is the passive one
	OnPassive func(ctx context.Context)

	// OnStop is called after the ActivePassive process stops
	OnStop func()

	close          chan struct{}
	clusterName    string
	timeToFailover time.Duration
	locker         *redislock.Client

	stateSetter StateSetter

	// current status of the failover (to show it in the readiness status)
	state   status
	stateMu sync.RWMutex
}

type ActivePassiveOption func(*ActivePassive) error

func WithCustomStateSetter(fn func(ctx context.Context, state string) error) ActivePassiveOption {
	return func(ap *ActivePassive) error {
		stateSetter, err := NewCustomStateSetter(fn)
		if err != nil {
			return fmt.Errorf("failed to create state setter: %w", err)
		}

		ap.stateSetter = stateSetter

		return nil
	}
}

func WithNoopStateSetter() ActivePassiveOption {
	return func(ap *ActivePassive) error {
		ap.stateSetter = &NoopStateSetter{}

		return nil
	}
}

func WithPodStateSetter() ActivePassiveOption {
	return func(ap *ActivePassive) error {
		stateSetter, err := NewPodStateSetter()
		if err != nil {
			return fmt.Errorf("failed to create pod state setter: %w", err)
		}

		ap.stateSetter = stateSetter

		return nil
	}
}

// NewActivePassive creates a new active passive cluster
// identified by the name. The time to fail over determines
// the frequency of checks performed against redis to
// keep the active state.
// NOTE: creating multiple ActivePassive in one process
// is not working correctly as there is only one readiness probe.
func NewActivePassive(clusterName string, timeToFailover time.Duration, client *redis.Client, opts ...ActivePassiveOption) (*ActivePassive, error) {
	activePassive := &ActivePassive{
		clusterName:    clusterName,
		timeToFailover: timeToFailover,
		locker:         redislock.New(client),
	}

	for _, opt := range opts {
		if err := opt(activePassive); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	if activePassive.stateSetter == nil {
		var err error

		// Default state setter uses the k8s api to set the state.
		activePassive.stateSetter, err = NewPodStateSetter()
		if err != nil {
			return nil, fmt.Errorf("failed to create default state setter: %w", err)
		}
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

	defer func() {
		if a.OnStop != nil {
			a.OnStop()
		}
	}()

	var lock *redislock.Lock

	// Ticker to try to refresh the lock's TTL before it expires
	tryRefreshLock := time.NewTicker(a.timeToFailover)

	// Ticker to try to acquire the lock if in passive or undefined state
	tryAcquireLock := time.NewTicker(500 * time.Millisecond)

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
		case <-tryAcquireLock.C:
			if a.getState() != ACTIVE {
				var err error

				lock, err = a.locker.Obtain(ctx, lockName, a.timeToFailover, &redislock.Options{
					RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(a.timeToFailover/3), 3),
				})
				if err != nil {
					if a.getState() != PASSIVE {
						logger.Info().Err(err).Msg("failed to obtain the lock; becoming passive...")
						a.becomePassive(ctx)
					}

					continue
				}

				logger.Debug().Msg("lock acquired; becoming active...")
				a.becomeActive(ctx)

				// Reset the refresh ticker to half of the time to failover
				tryRefreshLock.Reset(a.timeToFailover / 2)
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
			a.OnActive(ctx)
		}
	}
}

func (a *ActivePassive) becomePassive(ctx context.Context) {
	if a.setState(ctx, PASSIVE) {
		if a.OnPassive != nil {
			a.OnPassive(ctx)
		}
	}
}

func (a *ActivePassive) becomeUndefined(ctx context.Context) {
	a.setState(ctx, UNDEFINED)
}

// setState returns true if the state was set successfully
func (a *ActivePassive) setState(ctx context.Context, state status) bool {
	err := a.stateSetter.SetState(ctx, a.label(state))
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
