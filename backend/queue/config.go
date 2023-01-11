package queue

import (
	"log"
	"time"

	"github.com/caarlos0/env"
)

type config struct {
	HealthCheckResultTTL time.Duration `env:"RMQ_HEALTH_CHECK_RESULT_TTL" envDefault:"10s"`
	// HealthCheckPendingStateInterval represents the time between a queue becoming unhealthy and marking it as unhealthy.
	// Used to prevent a queue from becoming immediately unhealthy when a surge of deliveries occurs,
	// providing time to the service that uses this implementation to deal with the sudden increase in work without being
	// signaled as unhealthy.
	HealthCheckPendingStateInterval time.Duration `env:"RMQ_HEALTH_CHECK_PENDING_STATE_INTERVAL" envDefault:"1m"`
	MetricsRefreshInterval          time.Duration `env:"RMQ_METRICS_REFRESH_INTERVAL" envDefault:"10s"`
}

var cfg config

func init() {
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Failed to parse queue environment: %v", err)
	}
}
