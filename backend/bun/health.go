// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.

package bun

import (
	"context"
	"time"

	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
	"github.com/uptrace/bun"
)

// HealthCheck checks the state of a postgres connection. It must not be changed
// after it was registered as a health check.
type HealthCheck struct {
	state servicehealthcheck.ConnectionState
	db    *bun.DB
}

type healthcheck struct {
	bun.BaseModel

	OK bool `bun:"column:ok"`
}

// Init initializes the test table
func (h *HealthCheck) Init(ctx context.Context) error {
	_, err := h.db.NewCreateTable().Table(cfg.HealthCheckTableName).Model((*healthcheck)(nil)).Exec(ctx)

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
	if _, err := h.db.NewSelect().Column("1").Exec(ctx); err != nil {
		h.state.SetErrorState(err)
		return h.state.GetState()
	}

	hc := &healthcheck{OK: true}

	// writecheck - add Data to configured Table
	_, err := h.db.NewInsert().Model(hc).Exec(ctx)
	if err != nil {
		h.state.SetErrorState(err)
		return h.state.GetState()
	}

	// and while we're at it, check delete as well (so as not to clutter the database
	// because UPSERT is impractical here
	_, err = h.db.NewDelete().Exec(ctx)
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
	_, err := h.db.NewDropTable().Table(cfg.HealthCheckTableName).IfExists().Exec(ctx)

	return err
}
