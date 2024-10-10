package servicehealthcheck

import (
	"context"
	"errors"
)

var (
	_ Initializable = (*mockHealthCheck)(nil)
	_ HealthCheck   = (*mockHealthCheck)(nil)
)

type mockHealthCheck struct {
	initErr         bool
	healthCheckErr  bool
	healthCheckWarn bool
	name            string
}

func (t *mockHealthCheck) Init(_ context.Context) error {
	if t.initErr {
		return errors.New("initError")
	}

	return nil
}

func (t *mockHealthCheck) HealthCheck(_ context.Context) HealthCheckResult {
	if t.healthCheckErr {
		return HealthCheckResult{State: Err, Msg: "healthCheckErr"}
	}

	return HealthCheckResult{State: Ok}
}
