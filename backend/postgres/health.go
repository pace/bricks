// Copyright © 2024 by PACE Telematics GmbH. All rights reserved.

package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
	"github.com/uptrace/bun"
)

type queryExecutor interface {
	Exec(ctx context.Context, dest ...interface{}) (sql.Result, error)
}

// HealthCheck checks the state of a postgres connection. It must not be changed
// after it was registered as a health check.
type HealthCheck struct {
	state servicehealthcheck.ConnectionState

	createTableQueryExecutor queryExecutor
	deleteQueryExecutor      queryExecutor
	dropTableQueryExecutor   queryExecutor
	insertQueryExecutor      queryExecutor
	selectQueryExecutor      queryExecutor
}

type healthcheck struct {
	bun.BaseModel

	OK bool `bun:"column:ok"`
}

// NewHealthCheck creates a new HealthCheck instance.
func NewHealthCheck(db *bun.DB) *HealthCheck {
	return &HealthCheck{
		createTableQueryExecutor: db.NewCreateTable().Model((*healthcheck)(nil)).ModelTableExpr(cfg.HealthCheckTableName).IfNotExists(),
		deleteQueryExecutor:      db.NewDelete().ModelTableExpr(cfg.HealthCheckTableName).Where("TRUE"),
		dropTableQueryExecutor:   db.NewDropTable().ModelTableExpr(cfg.HealthCheckTableName).IfExists(),
		insertQueryExecutor:      db.NewInsert().ModelTableExpr(cfg.HealthCheckTableName).Model(&healthcheck{OK: true}),
		selectQueryExecutor:      db.NewRaw("SELECT 1;"),
	}
}

// Init initializes the test table
func (h *HealthCheck) Init(ctx context.Context) error {
	_, err := h.createTableQueryExecutor.Exec(ctx)
	return err
}

// HealthCheck performs the read test on the database. If enabled, it performs a
// write test as well.
func (h *HealthCheck) HealthCheck(ctx context.Context) servicehealthcheck.HealthCheckResult {
	if time.Since(h.state.LastChecked()) <= cfg.HealthCheckResultTTL {
		// the last result of the Health Check is still not outdated
		return h.state.GetState()
	}

	// Readcheck
	if _, err := h.selectQueryExecutor.Exec(ctx); err != nil {
		h.state.SetErrorState(err)
		return h.state.GetState()
	}

	// writecheck - add Data to configured Table
	if _, err := h.insertQueryExecutor.Exec(ctx); err != nil {
		h.state.SetErrorState(err)
		return h.state.GetState()
	}

	// and while we're at it, check delete as well (so as not to clutter the database
	// because UPSERT is impractical here
	if _, err := h.deleteQueryExecutor.Exec(ctx); err != nil {
		h.state.SetErrorState(err)
		return h.state.GetState()
	}

	// If no error occurred set the State of this Health Check to healthy
	h.state.SetHealthy()

	return h.state.GetState()
}

// CleanUp drops the test table.
func (h *HealthCheck) CleanUp(ctx context.Context) error {
	_, err := h.dropTableQueryExecutor.Exec(ctx)

	return err
}
