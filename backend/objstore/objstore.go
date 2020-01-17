// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/12 by Vincent Landgraf

// Package objstorage helps creating object storage connection pools
package objstore

import (
	"net/http"
	"time"

	"github.com/caarlos0/env"
	"github.com/minio/minio-go/v6"
	"github.com/pace/bricks/http/transport"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
	"github.com/pace/bricks/maintenance/log"
	"github.com/prometheus/client_golang/prometheus"
)

type config struct {
	Endpoint        string `env:"S3_ENDPOINT"`
	AccessKeyID     string `env:"S3_ACCESS_KEY_ID"`
	SecretAccessKey string `env:"S3_SECRET_ACCESS_KEY"`
	UseSSL          bool   `env:"S3_USE_SSL"`

	HealthCheckBucketName string        `env:"S3_HEALTH_CHECK_BUCKET_NAME" envDefault:"health-check"`
	HealthCheckObjectName string        `env:"S3_HEALTH_CHECK_OBJECT_NAME" envDefault:"latest.log"`
	HealthCheckResultTTL  time.Duration `env:"S3_HEALTH_CHECK_RESULT_TTL" envDefault:"2m"`
}

var (
	paceObjStoreTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pace_objstore_req_total",
			Help: "Collects stats about the number of object storage requests made",
		},
		[]string{"method", "bucket"},
	)
	paceObjStoreFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pace_objstore_req_failed",
			Help: "Collects stats about the number of object storage requests counterFailed",
		},
		[]string{"method", "bucket"},
	)
	paceObjStoreDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pace_objstore_req_duration_seconds",
			Help:    "Collect performance metrics for each method & bucket",
			Buckets: []float64{.1, .25, .5, 1, 2.5, 5, 10, 60},
		},
		[]string{"method", "bucket"},
	)
)

var cfg config

func init() {
	prometheus.MustRegister(paceObjStoreTotal)
	prometheus.MustRegister(paceObjStoreFailed)
	prometheus.MustRegister(paceObjStoreDurationSeconds)

	// parse log config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Failed to parse object storage environment: %v", err)
	}

	client, err := Client()
	if err != nil {
		log.Fatalf("Failed to create object storage client: %v", err)
	}
	servicehealthcheck.RegisterHealthCheck(&HealthCheck{
		Client: client,
	}, "objstore")
}

// Client with environment based configuration
func Client() (*minio.Client, error) {
	client, err := minio.New(cfg.Endpoint, cfg.AccessKeyID, cfg.SecretAccessKey, cfg.UseSSL)
	if err != nil {
		return nil, err
	}
	client.SetCustomTransport(newCustomTransport(cfg.Endpoint))
	return client, nil
}

// CustomClient with customized client
func CustomClient(endpoint string, opts *minio.Options) (*minio.Client, error) {
	client, err := minio.NewWithOptions(endpoint, opts)
	if err != nil {
		return nil, err
	}
	client.SetCustomTransport(newCustomTransport(cfg.Endpoint))
	return client, nil
}

func newCustomTransport(endpoint string) http.RoundTripper {
	rt := make([]transport.ChainableRoundTripper, 0)
	rt = append(rt, transport.NewDefaultTransportRoundTrippers()...)
	rt = append(rt, newMetricRoundTripper(endpoint))
	return transport.Chain(rt...)
}

func newMetricRoundTripper(endpoint string) *metricRoundTripper {
	return &metricRoundTripper{
		endpoint:      endpoint,
		counterTotal:  paceObjStoreTotal,
		counterFailed: paceObjStoreFailed,
		histogramDur:  paceObjStoreDurationSeconds,
	}
}
