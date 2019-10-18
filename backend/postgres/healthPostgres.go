// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/18 Charlotte Pröller

package postgres

import (
	"github.com/go-pg/pg"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
	"time"
)

type postgresHealthCheck struct {
	State *servicehealthcheck.ConnectionState
	pool  *pg.DB
}

func (h *postgresHealthCheck) InitHealthCheck() error {
	h.pool = ConnectionPool()
	h.State = servicehealthcheck.NewConnectionState()
	return nil
}

func (h *postgresHealthCheck) Name() string {
	return "postgres"
}

func (h *postgresHealthCheck) HealthCheck(currTime time.Time) (bool, error) {

	if currTime.Sub(h.State.LastCheck) > cfg.HealthMaxRequest {
		h.State.SetHealthy(currTime)
		//Readcheck
		_, errRead := h.pool.Exec(`SELECT 1; `)
		if errRead != nil {
			h.State.SetErrorState(errRead, currTime)
		} else {
			//writecheck - create Table if not exist and add Data
			_, errWrite := h.pool.Exec(`CREATE TABLE IF NOT EXISTS ` + cfg.HealthTableName + `(ok boolean);`)
			if errWrite == nil {
				_, errWrite = h.pool.Exec("INSERT INTO " + cfg.HealthTableName + "(ok) VALUES (true);")
			}
			if errWrite != nil {
				h.State.SetErrorState(errWrite, currTime)
			}
		}
	}
	return h.State.GetState()
}

func (h *postgresHealthCheck) CleanUp() error {
	_, err := h.pool.Exec("DROP TABLE IF EXISTS" + cfg.HealthTableName)
	return err

}
