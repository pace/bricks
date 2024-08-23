// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.

package postgres

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-pg/pg"
	"github.com/prometheus/client_golang/prometheus"
)

// ConnectionPoolMetrics is the metrics collector for postgres connection pools
// (pace_postgres_connection_pool_*). It is capable of running an observer that
// periodically gathers those stats.
type ConnectionPoolMetrics struct {
	poolMetrics   map[string]struct{}
	poolMetricsMx sync.Mutex

	hits       *prometheus.CounterVec
	misses     *prometheus.CounterVec
	timeouts   *prometheus.CounterVec
	totalConns *prometheus.GaugeVec
	idleConns  *prometheus.GaugeVec
	staleConns *prometheus.GaugeVec
}

// NewConnectionPoolMetrics returns a new metrics collector for postgres
// connection pools.
func NewConnectionPoolMetrics() *ConnectionPoolMetrics {
	m := ConnectionPoolMetrics{
		poolMetrics: map[string]struct{}{},
		hits: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "pace_postgres_connection_pool_hits",
				Help: "Collects number of times free connection was found in the pool",
			},
			[]string{"database", "pool"},
		),
		misses: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "pace_postgres_connection_pool_misses",
				Help: "Collects number of times free connection was NOT found in the pool",
			},
			[]string{"database", "pool"},
		),
		timeouts: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "pace_postgres_connection_pool_timeouts",
				Help: "Collects number of times a wait timeout occurred",
			},
			[]string{"database", "pool"},
		),
		totalConns: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pace_postgres_connection_pool_total_conns",
				Help: "Collects number of total connections in the pool",
			},
			[]string{"database", "pool"},
		),
		idleConns: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pace_postgres_connection_pool_idle_conns",
				Help: "Collects number of idle connections in the pool",
			},
			[]string{"database", "pool"},
		),
		staleConns: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pace_postgres_connection_pool_stale_conns",
				Help: "Collects number of stale connections removed from the pool",
			},
			[]string{"database", "pool"},
		),
	}
	return &m
}

// The metrics implement the prometheus collector methods. This allows to
// register them directly with a registry.
var _ prometheus.Collector = (*ConnectionPoolMetrics)(nil)

// Describe descibes all the embedded prometheus metrics.
func (m *ConnectionPoolMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.hits.Describe(ch)
	m.misses.Describe(ch)
	m.timeouts.Describe(ch)
	m.totalConns.Describe(ch)
	m.idleConns.Describe(ch)
	m.staleConns.Describe(ch)
}

// Collect collects all the embedded prometheus metrics.
func (m *ConnectionPoolMetrics) Collect(ch chan<- prometheus.Metric) {
	m.hits.Collect(ch)
	m.misses.Collect(ch)
	m.timeouts.Collect(ch)
	m.totalConns.Collect(ch)
	m.idleConns.Collect(ch)
	m.staleConns.Collect(ch)
}

// ObserveRegularly starts observing the given postgres pool. The provided pool
// name must be unique as it distinguishes multiple pools. The pool name is
// exposed as the "pool" label in the metrics. The metrics are collected once
// per minute for as long as the passed context is valid.
func (m *ConnectionPoolMetrics) ObserveRegularly(ctx context.Context, db *pg.DB, poolName string) error {
	trigger := make(chan chan<- struct{})
	if err := m.ObserveWhenTriggered(trigger, db, poolName); err != nil {
		return err
	}

	// Trigger once a minute until context is cancelled. In the following
	// goroutine we create a ticker that writes to a channel every minute. If
	// this happens we write to the trigger channel and that will trigger
	// observing the metrics. Both channel operations are blocking which is why
	// we have to check the context two times. So that the goroutine doesn't
	// stick around forever which would prevent the garbage collection from
	// cleaning up the related resources.
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer close(trigger)
		for {
			select {
			case <-ticker.C:
				select {
				// The trigger channel allows passing another channel if we
				// wanted to get notified when observing the metrics is done.
				// But we don't, so we just pass nil.
				case trigger <- nil:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// ObserveWhenTriggered starts observing the given postgres pool. The pool name
// behaves as decribed for the ObserveRegularly method. The metrics are observed
// for every emitted value from the trigger channel. The trigger channel allows
// passing a response channel that will be closed once the metrics were
// collected. It is also possible to pass nil. You should close the trigger
// channel when done to allow cleaning up.
func (m *ConnectionPoolMetrics) ObserveWhenTriggered(trigger <-chan chan<- struct{}, db *pg.DB, poolName string) error {
	// check that pool name is unique
	m.poolMetricsMx.Lock()
	defer m.poolMetricsMx.Unlock()
	if _, ok := m.poolMetrics[poolName]; ok {
		return fmt.Errorf("invalid pool name: %q: %w", poolName, ErrNotUnique)
	}
	m.poolMetrics[poolName] = struct{}{}

	// start goroutine
	go m.gatherConnectionPoolMetrics(trigger, db, poolName)
	return nil
}

func (m *ConnectionPoolMetrics) gatherConnectionPoolMetrics(trigger <-chan chan<- struct{}, db *pg.DB, poolName string) {
	// prepare labels for all stats
	opts := db.Options()
	labels := prometheus.Labels{
		"database": opts.Addr + "/" + opts.Database,
		"pool":     poolName,
	}

	// keep previous stats for the counters
	var prevStats pg.PoolStats

	// collect all the pool stats whenever triggered
	for done := range trigger {
		stats := db.PoolStats()
		// counters
		m.hits.With(labels).Add(float64(stats.Hits - prevStats.Hits))
		m.misses.With(labels).Add(float64(stats.Misses - prevStats.Misses))
		m.timeouts.With(labels).Add(float64(stats.Timeouts - prevStats.Timeouts))
		// gauges
		m.totalConns.With(labels).Set(float64(stats.TotalConns))
		m.idleConns.With(labels).Set(float64(stats.IdleConns))
		m.staleConns.With(labels).Set(float64(stats.StaleConns))
		// inform caller that we are done
		if done != nil {
			close(done)
		}
		prevStats = *stats
	}
}
