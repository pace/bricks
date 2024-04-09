// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/18 by Charlotte Pröller

package postgres

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10/orm"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
)

// HealthCheck checks the state of a postgres connection. It must not be changed
// after it was registered as a health check.
type HealthCheck struct {
	state servicehealthcheck.ConnectionState
	Pool  postgresQueryExecutor
}

type postgresQueryExecutor interface {
	Exec(ctx context.Context, query interface{}, params ...interface{}) (res orm.Result, err error)
}

// Init initializes the test table
func (h *HealthCheck) Init(ctx context.Context) error {
	_, errWrite := h.Pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS `+cfg.HealthCheckTableName+`(ok boolean);`)
	return errWrite
}

// HealthCheck performs the read test on the database. If enabled, it performs a
// write test as well.
func (h *HealthCheck) HealthCheck(ctx context.Context) servicehealthcheck.HealthCheckResult {
	if time.Since(h.state.LastChecked()) <= cfg.HealthCheckResultTTL {
		// the last result of the Health Check is still not outdated
		return h.state.GetState()
	}

	// Readcheck
	if _, err := h.Pool.Exec(ctx, `SELECT 1;`); err != nil {
		h.state.SetErrorState(err)
		return h.state.GetState()
	}
	// writecheck - add Data to configured Table
	_, err := h.Pool.Exec(ctx, "INSERT INTO "+cfg.HealthCheckTableName+"(ok) VALUES (true);")
	if err != nil {
		h.state.SetErrorState(err)
		return h.state.GetState()
	}
	// and while we're at it, check delete as well (so as not to clutter the database
	// because UPSERT is impractical here
	_, err = h.Pool.Exec(ctx, "DELETE FROM "+cfg.HealthCheckTableName+";")
	if err != nil {
		h.state.SetErrorState(err)
		return h.state.GetState()
	}
	// If no error occurred set the State of this Health Check to healthy
	h.state.SetHealthy()
	return h.state.GetState()

}

// CleanUp drops the test table.
func (h *HealthCheck) CleanUp(ctx context.Context) error {
	_, err := h.Pool.Exec(ctx, "DROP TABLE IF EXISTS "+cfg.HealthCheckTableName)
	return err
}
