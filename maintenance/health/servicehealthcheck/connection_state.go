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
	isHealthy bool
	err       error
	m         sync.Mutex
}

func (cs *ConnectionState) setConnectionState(healthy bool, err error) {
	cs.m.Lock()
	defer cs.m.Unlock()
	cs.isHealthy = healthy
	cs.err = err
	cs.lastCheck = time.Now()
}

// SetErrorState sets the state to not healthy.
func (cs *ConnectionState) SetErrorState(err error) {
	cs.setConnectionState(err == nil, err)
}

// SetHealthy sets the state to healthy.
func (cs *ConnectionState) SetHealthy() {
	cs.setConnectionState(true, nil)
}

// GetState returns the current state. That is whether the check is healthy or
// the error occured.
func (cs *ConnectionState) GetState() (bool, error) {
	cs.m.Lock()
	defer cs.m.Unlock()
	return cs.isHealthy, cs.err
}

// LastChecked returns the time that the state was last updated or confirmed.
func (cs *ConnectionState) LastChecked() time.Time {
	cs.m.Lock()
	defer cs.m.Unlock()
	return cs.lastCheck
}
