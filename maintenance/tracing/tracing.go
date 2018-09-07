// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/07 by Vincent Landgraf

package tracing

import (
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics/prometheus"
	"lab.jamit.de/pace/go-microservice/maintenance/log"
)

func init() {
	cfg, err := config.FromEnv()
	if cfg.ServiceName == "" {
		log.Warn("Using Jaeger noop tracer since no JAEGER_SERVICE_NAME is present")
		return
	}

	if err != nil {
		log.Warnf("Unable to load Jaeger config from ENV: %v", err)
		return
	}

	tracer, _, err := cfg.NewTracer(
		config.Metrics(prometheus.New()),
	)
	opentracing.SetGlobalTracer(tracer)
	if err != nil {
		log.Fatal(err)
	}
}
