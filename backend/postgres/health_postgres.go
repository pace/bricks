// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/18 by Charlotte Pröller

package postgres

import (
	"time"

	"github.com/go-pg/pg"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
)

// HealthCheck checks the state of a postgres connection. It must not be changed
// after it was registered as a health check.
type HealthCheck struct {
	state      servicehealthcheck.ConnectionState
	Pool       *pg.DB
	CheckWrite bool
}

// Init initialises the test table if the write check is enabled.
func (h *HealthCheck) Init() error {
	if h.CheckWrite {
		_, errWrite := h.Pool.Exec(`CREATE TABLE IF NOT EXISTS ` + cfg.HealthTableName + `(ok boolean);`)
		return errWrite
	}
	return nil
}

// HealthCheck performs the read test on the database. If enabled, it performs a
// write test as well.
func (h *HealthCheck) HealthCheck() (bool, error) {
	if time.Since(h.state.LastChecked()) <= cfg.HealthMaxRequest {
		// the last result of the Health Check is still not outdated
		return h.state.GetState()
	}

	// Readcheck
	if _, err := h.Pool.Exec(`SELECT 1;`); err != nil {
		h.state.SetErrorState(err)
		return h.state.GetState()
	}
	// writecheck - add Data to configured Table
	if h.CheckWrite {
		_, err := h.Pool.Exec("INSERT INTO " + cfg.HealthTableName + "(ok) VALUES (true);")
		if err != nil {
			h.state.SetErrorState(err)
			return h.state.GetState()
		}
	}
	// If no error occurred set the State of this Health Check to healthy
	h.state.SetHealthy()
	return h.state.GetState()

}

// CleanUp drops the test table.
func (h *HealthCheck) CleanUp() error {
	if h.CheckWrite {
		_, err := h.Pool.Exec("DROP TABLE IF EXISTS " + cfg.HealthTableName)
		return err
	}
	return nil
}
