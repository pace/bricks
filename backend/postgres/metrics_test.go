// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/11/21 by Marius Neugebauer

package postgres_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/pace/bricks/backend/postgres"
)

func ExampleConnectionPoolMetrics() {
	myDB := ConnectionPool()

	// collect stats about my db every minute
	metrics := NewConnectionPoolMetrics()
	if err := metrics.ObserveRegularly(context.Background(), myDB, "my_db"); err != nil {
		panic(err)
	}
	prometheus.MustRegister(metrics)
}

func TestIntegrationConnectionPoolMetrics(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	// prepare connection pool with metrics
	metricsRegistry := prometheus.NewRegistry()
	metrics := NewConnectionPoolMetrics()
	metricsRegistry.MustRegister(metrics)
	db := ConnectionPool()
	trigger := make(chan chan<- struct{})
	err := metrics.ObserveWhenTriggered(trigger, db, "test")
	require.NoError(t, err)
	// collect some metrics
	if _, err := db.Exec(`SELECT 1;`); err != nil {
		t.Fatalf("could not query postgres database: %s", err)
	}
	whenDone := make(chan struct{})
	select {
	case trigger <- whenDone:
	case <-time.After(time.Second):
		t.Fatal("did not start collecting metrics after 1s")
	}
	select {
	case <-whenDone:
	case <-time.After(time.Second):
		t.Fatal("metrics were not collected after 1s")
	}
	// query metrics
	resp := httptest.NewRecorder()
	handler := promhttp.HandlerFor(metricsRegistry, promhttp.HandlerOpts{})
	handler.ServeHTTP(resp, httptest.NewRequest("GET", "/metrics", nil))
	body := resp.Body.String()
	assert.Regexp(t, `pace_postgres_connection_pool_hits.*?\Wpool="test"\W`, body)
	assert.Regexp(t, `pace_postgres_connection_pool_misses.*?\Wpool="test"\W`, body)
	assert.Regexp(t, `pace_postgres_connection_pool_timeouts.*?\Wpool="test"\W`, body)
	assert.Regexp(t, `pace_postgres_connection_pool_total_conns.*?\Wpool="test"\W`, body)
	assert.Regexp(t, `pace_postgres_connection_pool_idle_conns.*?\Wpool="test"\W`, body)
	assert.Regexp(t, `pace_postgres_connection_pool_stale_conns.*?\Wpool="test"\W`, body)
}

// Tests that the NewConnectionPoolMetrics don't allow registering pools using
// the same pool name.
func TestIntegrationConnectionPoolMetrics_duplicatePoolName(t *testing.T) {
	metrics := NewConnectionPoolMetrics()
	// register first with name "test"
	err := metrics.ObserveRegularly(context.Background(), ConnectionPool(), "test")
	require.NoError(t, err)
	// registering second with name "test" fails
	err = metrics.ObserveRegularly(context.Background(), ConnectionPool(), "test")
	assert.True(t, errors.Is(err, ErrNotUnique))
}
