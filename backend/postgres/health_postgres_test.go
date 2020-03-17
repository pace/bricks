// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/18 by Charlotte Pröller

package postgres

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-pg/pg/orm"
	http2 "github.com/pace/bricks/http"
	"github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
	"github.com/pace/bricks/maintenance/log"
	"github.com/stretchr/testify/require"
)

func setup() *http.Response {
	r := http2.Router()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health/check", nil)
	r.ServeHTTP(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	return resp
}

func TestIntegrationHealthCheck(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	resp := setup()
	if resp.StatusCode != 200 {
		t.Errorf("Expected /health/check to respond with 200, got: %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if !strings.Contains(string(data[:]), "postgresdefault        OK") {
		t.Errorf("Expected /health/check to return OK, got: %q", string(data[:]))
	}
}

type testPool struct {
	err error
}

func (t *testPool) Exec(ctx context.Context, query interface{}, params ...interface{}) (res orm.Result, err error) {
	return nil, t.err
}

func TestHealthCheckCaching(t *testing.T) {
	ctx := context.Background()

	// set the TTL to a minute because this is long enough to test that the result is cached
	cfg.HealthCheckResultTTL = time.Minute
	requiredErr := errors.New("TestHealthCheckCaching")
	pool := &testPool{err: requiredErr}
	h := &HealthCheck{Pool: pool}
	res := h.HealthCheck(ctx)
	// get the error for the first time
	require.Equal(t, servicehealthcheck.Err, res.State)
	require.Equal(t, "TestHealthCheckCaching", res.Msg)
	res = h.HealthCheck(ctx)
	pool.err = nil
	// getting the cached error
	require.Equal(t, servicehealthcheck.Err, res.State)
	require.Equal(t, "TestHealthCheckCaching", res.Msg)
	// Resetting the TTL to get a uncached result
	cfg.HealthCheckResultTTL = 0
	res = h.HealthCheck(ctx)
	require.Equal(t, servicehealthcheck.Ok, res.State)
	require.Equal(t, "", res.Msg)
}
