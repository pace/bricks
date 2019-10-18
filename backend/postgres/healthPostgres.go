package postgres

import (
	"github.com/go-pg/pg"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
	"sync"
	"time"
)

type postgresHealthCheck struct {
	State *servicehealthcheck.ConnectionState
	pool  *pg.DB
}

func (h *postgresHealthCheck) InitHealthCheck() error {
	h.pool = ConnectionPool()
	h.State = &servicehealthcheck.ConnectionState{M: sync.Mutex{}}
	return nil
}

func (h *postgresHealthCheck) Name() string {
	return "postgres"
}

func (h *postgresHealthCheck) HealthCheck(currTime time.Time) error {
	h.State.M.Lock()
	defer h.State.M.Unlock()
	if currTime.Sub(h.State.LastCheck) > cfg.HealthMaxRequest {
		//Readcheck
		_, errRead := h.pool.Exec(`SELECT 1; `)
		if errRead != nil {
			h.State.IsHealthy = false
			h.State.LastCheck = currTime
			h.State.Err = errRead
			//if the readcheck failes we don't have to try the write check, it will not work
			return h.State.Err
		}
		//writecheck - create Table if not exist and add Data
		_, errWrite := h.pool.Exec(`CREATE TABLE IF NOT EXISTS ` + cfg.HealthTableName + `(ok boolean);`)
		if errWrite == nil {
			_, errWrite = h.pool.Exec("INSERT INTO " + cfg.HealthTableName + "(ok) VALUES (true);")
		}
		h.State.IsHealthy = errWrite == nil
		h.State.Err = errWrite
		h.State.LastCheck = currTime
	}
	return h.State.Err
}

func (h *postgresHealthCheck) CleanUp() error {
	_, err := h.pool.Exec("DROP TABLE IF EXISTS" + cfg.HealthTableName)
	return err

}
