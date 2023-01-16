package queue

import (
	"context"
	"testing"
	"time"

	"github.com/pace/bricks/maintenance/log"
	"github.com/stretchr/testify/assert"
)

func TestIntegrationHealthCheck(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	ctx := log.WithContext(context.Background())
	cfg.HealthCheckPendingStateInterval = time.Second * 2
	q1, err := NewQueue("integrationTestTasks", 1)
	assert.NoError(t, err)
	err = q1.Publish("nothing here")
	assert.NoError(t, err)

	time.Sleep(time.Second)
	check := &HealthCheck{IgnoreInterval: true}
	res := check.HealthCheck(ctx)
	if res.State != "OK" {
		t.Errorf("Expected health check to be OK for a non-full queue: state %s, message: %s", res.State, res.Msg)
	}

	err = q1.Publish("nothing here either")
	assert.NoError(t, err)

	// queue health started pending
	res = check.HealthCheck(ctx)
	if res.State != "OK" {
		t.Errorf("Expected health check to be OK")
	}
	// queue health pending
	time.Sleep(time.Second)
	res = check.HealthCheck(ctx)
	if res.State != "OK" {
		t.Errorf("Expected health check to be OK")
	}
	// queue health no longer pending
	time.Sleep(time.Second * 2)
	res = check.HealthCheck(ctx)
	if res.State == "OK" {
		t.Errorf("Expected health check to be ERR for a full queue")
	}

	_, _ = q1.PurgeReady()
	// queue health back to OK
	res = check.HealthCheck(ctx)
	if res.State != "OK" {
		t.Errorf("Expected health check to be OK")
	}

	err = q1.Publish("nothing here")
	assert.NoError(t, err)
	err = q1.Publish("nothing here either")
	assert.NoError(t, err)
	// queue health pending again
	res = check.HealthCheck(ctx)
	if res.State != "OK" {
		t.Errorf("Expected health check to be OK")
	}

	_, _ = q1.PurgeReady()
}
