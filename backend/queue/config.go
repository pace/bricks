package queue

import (
	"log"
	"time"

	"github.com/caarlos0/env"
)

type config struct {
	HealthCheckResultTTL   time.Duration `env:"RMQ_HEALTH_CHECK_RESULT_TTL" envDefault:"10s"`
	MetricsRefreshInterval time.Duration `env:"RMQ_METRICS_REFRESH_INTERVAL" envDefault:"10s"`
}

var cfg config

func init() {
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Failed to parse queue environment: %v", err)
	}
}
