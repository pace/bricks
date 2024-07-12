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

// ActivePassive implements a failover mechanism that allows
// to deploy a service multiple times but ony one will accept
// traffic by using the label selector of kubernetes.
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

	// access to the kubernetes api
	client *k8sapi.Client

	// current status of the failover (to show it in the readiness status)
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
// and OnPassive callbacks in case the election toke place.
// Will handle panic safely and therefore can be directly called
// with go.
func (a *ActivePassive) Run(ctx context.Context) error {
	defer errors.HandleWithCtx(ctx, "activepassive failover handler")

	lockName := "activepassive:lock:" + a.clusterName
	logger := log.Ctx(ctx).With().Str("failover", lockName).Logger()
	ctx = logger.WithContext(ctx)

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
				err := lock.Refresh(ctx, a.timeToFailover, &redislock.Options{
					RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(a.timeToFailover/3), 3),
				})
				if err != nil {
					logger.Debug().Err(err).Msg("failed to refresh")
					a.becomeUndefined(ctx)
				}
			}
		case <-retry.C:
			// try to acquire the lock, as we are not the active
			if a.getState() != ACTIVE {
				var err error
				lock, err = a.locker.Obtain(ctx, lockName, a.timeToFailover, &redislock.Options{
					RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(a.timeToFailover/3), 3),
				})
				if err != nil {
					// we became passive, trigger callback
					if a.getState() != PASSIVE {
						logger.Debug().Err(err).Msg("becoming passive")
						a.becomePassive(ctx)
					}

					continue
				}

				// lock acquired
				logger.Debug().Msg("becoming active")
				a.becomeActive(ctx)

				// we are active, renew if required
				d, err := lock.TTL(ctx)
				if err != nil {
					logger.Debug().Err(err).Msg("failed to get TTL")
				}
				if d == 0 {
					// TTL seems to be expired, retry to get lock or become
					// passive in next iteration
					logger.Debug().Msg("ttl expired")
					a.becomeUndefined(ctx)
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
