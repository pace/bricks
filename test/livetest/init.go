// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package livetest

import (
	"log"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/prometheus/client_golang/prometheus"
)

type config struct {
	Interval    time.Duration `env:"PACE_LIVETEST_INTERVAL" envDefault:"1h"`
	ServiceName string        `env:"JAEGER_SERVICE_NAME" envDefault:"pace-bricks"`
}

var (
	paceLivetestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pace_livetest_total",
			Help: "Collects stats about the number of live tests made",
		},
		[]string{"service", "result"},
	)
	paceLivetestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pace_livetest_duration_seconds",
			Help:    "Collect performance metrics for each live test",
			Buckets: []float64{.1, .25, .5, 1, 2.5, 5, 10, 60},
		},
		[]string{"service"},
	)
)

var cfg config

func init() {
	prometheus.MustRegister(paceLivetestTotal)
	prometheus.MustRegister(paceLivetestDurationSeconds)

	// parse log config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Failed to parse livetest environment: %v", err)
	}
}
