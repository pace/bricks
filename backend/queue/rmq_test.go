package queue_test

import (
	"context"
	"testing"
	"time"

	"github.com/pace/bricks/backend/queue"
	"github.com/pace/bricks/maintenance/log"
	"github.com/stretchr/testify/assert"
)

func TestIntegrationHealthCheck(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	ctx := log.WithContext(context.Background())
	q1, err := queue.NewQueue("integrationTestTasks", 1)
	assert.NoError(t, err)
	err = q1.Publish("nothing here")
	assert.NoError(t, err)

	time.Sleep(time.Second)
	check := &queue.HealthCheck{IgnoreInterval: true}
	res := check.HealthCheck(ctx)
	if res.State != "OK" {
		t.Errorf("Expected health check to be OK for a non-full queue: state %s, message: %s", res.State, res.Msg)
	}

	err = q1.Publish("nothing here either")
	assert.NoError(t, err)

	res = check.HealthCheck(ctx)
	if res.State == "OK" {
		t.Errorf("Expected health check to be ERR for a full queue")
	}
}
