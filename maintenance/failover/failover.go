// Copyright © 2022 by PACE Telematics GmbH. All rights reserved.
// Created at 2022/01/20 by Vincent Landgraf

package failover

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bsm/redislock"
	"github.com/caarlos0/env"
	"github.com/go-redis/redis/v7"
	"github.com/pace/bricks/http/transport"
	"github.com/pace/bricks/maintenance/errors"
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

const Label = "github.com.pace.bricks.activepassive"

// Config gathers the required kubernetes system configuration to use the
// kubernetes API
type Config struct {
	Host          string `env:"KUBERNETES_SERVICE_HOST" envDefault:"localhost"`
	Port          int    `env:"KUBERNETES_PORT_443_TCP_PORT" envDefault:"433"`
	NamespaceFile string `env:"KUBERNETES_NAMESPACE_FILE" envDefault:"/run/secrets/kubernetes.io/serviceaccount/namespace"`
	CACertFile    string `env:"KUBERNETES_API_CA_FILE" envDefault:"/run/secrets/kubernetes.io/serviceaccount/ca.crt"`
	TokenFile     string `env:"KUBERNETES_API_TOKEN_FILE" envDefault:"/run/secrets/kubernetes.io/serviceaccount/token"`
}

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

	Client *http.Client

	close          chan struct{}
	clusterName    string
	timeToFailover time.Duration
	locker         *redislock.Client
	cfg            Config
	podname        string
	Namespace      string
	CACert         []byte
	Token          string

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
	ap := &ActivePassive{
		clusterName:    clusterName,
		timeToFailover: timeToFailover,
		locker:         redislock.New(client),
		Client:         &http.Client{},
	}
	health.SetCustomReadinessCheck(ap.Handler)

	// lookup hostname (for pod update)
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	ap.podname = hostname

	// parse environment including secrets mounted by kubernetes
	err = env.Parse(&ap.cfg)
	if err != nil {
		return nil, err
	}

	caData, err := os.ReadFile(ap.cfg.CACertFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %v", ap.cfg.CACertFile, err)
	}
	ap.CACert = []byte(strings.TrimSpace(string(caData)))

	namespaceData, err := os.ReadFile(ap.cfg.NamespaceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %v", ap.cfg.NamespaceFile, err)
	}
	ap.Namespace = strings.TrimSpace(string(namespaceData))

	tokenData, err := os.ReadFile(ap.cfg.TokenFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %v", ap.cfg.CACertFile, err)
	}
	ap.Token = strings.TrimSpace(string(tokenData))

	// add kubernetes api server cert
	chain := transport.NewDefaultTransportChain()
	pool := x509.NewCertPool()
	ok := pool.AppendCertsFromPEM(ap.CACert)
	if !ok {
		return nil, fmt.Errorf("failed to load kubernetes ca cert")
	}
	chain.Final(&http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: pool,
		},
	})
	ap.Client.Transport = chain

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
				err := lock.Refresh(a.timeToFailover, &redislock.Options{
					RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(a.timeToFailover/3), 3),
				})
				if err != nil {
					logger.Debug().Err(err).Msgf("failed to refresh")
					a.becomeUndefined(ctx)
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
						a.becomePassive(ctx)
					}

					continue
				}

				// lock acquired
				logger.Debug().Msg("becoming active")
				a.becomeActive(ctx)

				// we are active, renew if required
				d, err := lock.TTL()
				if err != nil {
					logger.Debug().Err(err).Msgf("failed to get TTL %q")
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
	a.setState(ACTIVE)
	a.updateLabel(ctx, ACTIVE)
	if a.OnActive != nil {
		a.OnActive()
	}
}

func (a *ActivePassive) becomePassive(ctx context.Context) {
	a.setState(PASSIVE)
	a.updateLabel(ctx, PASSIVE)
	if a.OnPassive != nil {
		a.OnPassive()
	}
}

func (a *ActivePassive) becomeUndefined(ctx context.Context) {
	a.setState(UNDEFINED)
	a.updateLabel(ctx, UNDEFINED)
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

func (a *ActivePassive) updateLabel(ctx context.Context, s status) {
start:
	pr := patchRequest{
		{
			Op:    "add",
			Path:  "/metadata/labels/" + Label,
			Value: a.label(s),
		},
	}
	data, err := json.Marshal(pr)
	if err != nil {
		panic(err)
	}

	url := fmt.Sprintf("https://%s:%d/api/v1/namespaces/%s/pods/%s",
		a.cfg.Host, a.cfg.Port, a.Namespace, a.podname)

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(data))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json-patch+json")
	req.Header.Set("Authorization", "Bearer "+a.Token)

	resp, err := a.Client.Do(req)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("failed to do api request")
		time.Sleep(time.Second)
		goto start
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body) // nolint: errcheck
		log.Ctx(ctx).Debug().Msgf("failed to do api request, due to: %s", string(body))
	}
}

type patchRequest []struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}
