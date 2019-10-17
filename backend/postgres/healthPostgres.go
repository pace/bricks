package postgres

import (
	"github.com/go-pg/pg"
	"sync"
	"time"
)

type connectionState struct {
	moment    time.Time
	isHealthy bool
	err       error
	m         sync.Mutex
}

type postgresHealthCheck struct {
	State connectionState
	pool  *pg.DB
}

func (h postgresHealthCheck) InitHealthcheck() error {
	h.pool = ConnectionPool()
	h.State = connectionState{m: sync.Mutex{}}
	return nil
}

func (h postgresHealthCheck) Name() string {
	return "postgres"
}

func (h postgresHealthCheck) HealthCheck(currTime time.Time) error {
	h.State.m.Lock()
	defer h.State.m.Unlock()
	if currTime.Sub(h.State.moment).Seconds() > float64(cfg.HealthMaxRequestInSec) {
		//Readcheck
		errRead := h.pool.Select("1")
		if errRead != nil {
			h.State.isHealthy = false
			h.State.moment = currTime
			h.State.err = errRead
			//if the readcheck failes we don't have to try the write check, it will not work
			return h.State.err
		}
		//writecheck - create Table if not exist and add Data
		_, errWrite := h.pool.Exec(`CREATE TABLE IF NOT EXISTS ` + cfg.HealthTableName + `(ok boolean);`)
		if errWrite == nil {
			_, errWrite = h.pool.Exec("INSERT INTO " + cfg.HealthTableName + "(ok) VALUES (true);")
		}
		h.State.isHealthy = errWrite == nil
		h.State.err = errWrite
		h.State.moment = currTime
	}
	return h.State.err
}

func (h postgresHealthCheck) CleanUp() error {
	_, err := h.pool.Exec("DROP TABLE IF EXISTS" + cfg.HealthTableName)
	return err

}
