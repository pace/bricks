// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/18 Charlotte Pröller

package postgres

import (
	"github.com/go-pg/pg"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
	"time"
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

	return nil
}

func (h *PgHealthCheck) HealthCheck() (bool, error) {
	currTime := time.Now()
	if currTime.Sub(h.State.LastCheck) > cfg.HealthMaxRequest {
		h.State.SetHealthy(currTime)
		//Readcheck
		_, errRead := h.Pool.Exec(`SELECT 1; `)
		if errRead != nil {
			h.State.SetErrorState(errRead, currTime)
		} else {
			//writecheck - create Table if not exist and add Data
			_, errWrite := h.Pool.Exec(`CREATE TABLE IF NOT EXISTS ` + cfg.HealthTableName + `(ok boolean);`)
			if errWrite == nil {
				_, errWrite = h.Pool.Exec("INSERT INTO " + cfg.HealthTableName + "(ok) VALUES (true);")
			}
			if errWrite != nil {
				h.State.SetErrorState(errWrite, currTime)
			}
		}
	}
	return h.State.GetState()
}

func (h *PgHealthCheck) CleanUp() error {
	_, err := h.Pool.Exec("DROP TABLE IF EXISTS " + cfg.HealthTableName)
	return err

}
