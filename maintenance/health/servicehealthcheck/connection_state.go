// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/18 Charlotte Pröller

package servicehealthcheck

import (
	"sync"
	"time"
)

// ConnectionState caches the result of health checks. It is concurrency-safe.
type ConnectionState struct {
	lastCheck time.Time
	result    HealthCheckResult
	m         sync.Mutex
}

func (cs *ConnectionState) setConnectionState(result HealthCheckResult) {
	cs.m.Lock()
	defer cs.m.Unlock()
	cs.result = result
	cs.lastCheck = time.Now()
}

// SetErrorState sets the state to not healthy.
func (cs *ConnectionState) SetErrorState(err error) {
	cs.setConnectionState(HealthCheckResult{State: Err, Msg: err.Error()})
}

// SetHealthy sets the state to healthy.
func (cs *ConnectionState) SetHealthy() {
	cs.setConnectionState(HealthCheckResult{State: Ok, Msg: ""})
}

// GetState returns the current state. That is whether the check is healthy or
// the error occurred.
func (cs *ConnectionState) GetState() HealthCheckResult {
	cs.m.Lock()
	defer cs.m.Unlock()
	return cs.result
}

// LastChecked returns the time that the state was last updated or confirmed.
func (cs *ConnectionState) LastChecked() time.Time {
	cs.m.Lock()
	defer cs.m.Unlock()
	return cs.lastCheck
}
