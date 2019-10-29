// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/18 by Charlotte Pröller

package postgres

import (
	"time"

	"github.com/go-pg/pg"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
)

type PgHealthCheck struct {
	State *servicehealthcheck.ConnectionState
	Pool  *pg.DB
}

func (h *PgHealthCheck) InitHealthCheck() error {
	if h.Pool == nil {
		h.Pool = ConnectionPool()
	}
	h.State = servicehealthcheck.NewConnectionState()
	_, errWrite := h.Pool.Exec(`CREATE TABLE IF NOT EXISTS ` + cfg.HealthTableName + `(ok boolean);`)
	return errWrite
}

func (h *PgHealthCheck) HealthCheck() (bool, error) {
	currTime := time.Now()
	if currTime.Sub(h.State.LastCheck) <= cfg.HealthMaxRequest {
		// the last result of the Health Check is still not outdated
		return h.State.GetState()
	}

	// Readcheck
	_, errRead := h.Pool.Exec(`SELECT 1; `)
	if errRead != nil {
		h.State.SetErrorState(errRead, currTime)
		return h.State.GetState()
	}
	// writecheck - add Data to configured Table
	_, errWrite := h.Pool.Exec("INSERT INTO " + cfg.HealthTableName + "(ok) VALUES (true);")
	if errWrite != nil {
		h.State.SetErrorState(errWrite, currTime)
		return h.State.GetState()
	}
	// If no error occurred set the State of this Health Check to healthy
	h.State.SetHealthy(currTime)
	return h.State.GetState()

}

func (h *PgHealthCheck) CleanUp() error {
	_, err := h.Pool.Exec("DROP TABLE IF EXISTS " + cfg.HealthTableName)
	return err

}
