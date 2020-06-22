package queue_test

import (
	"context"
	"testing"

	"github.com/pace/bricks/backend/queue"
)

func TestIntegrationHealthCheck(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	q1 := queue.NewQueue("integrationTestTasks", 1)

	check := &queue.HealthCheck{}
	res := check.HealthCheck(context.Background())
	if res.State != "OK" {
		t.Errorf("Expected health check to be OK for an empty queue")
	}

	q1.Publish("nothing here")
	q1.Publish("nothing here either")

	res = check.HealthCheck(context.Background())
	if res.State == "OK" {
		t.Errorf("Expected health check to be ERR for a full queue")
	}
}
