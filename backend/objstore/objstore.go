package objstore

import (
	"net/http"
	"sync"
	"time"

	"github.com/caarlos0/env"
	"github.com/minio/minio-go/v6"
	"github.com/minio/minio-go/v6/pkg/credentials"

	"github.com/pace/bricks/http/transport"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
	"github.com/pace/bricks/maintenance/log"
)

type config struct {
	Endpoint        string `env:"S3_ENDPOINT" envDefault:"s3.amazonaws.com"`
	AccessKeyID     string `env:"S3_ACCESS_KEY_ID"`
	Region          string `env:"S3_REGION" envDefault:"us-east-1"`
	SecretAccessKey string `env:"S3_SECRET_ACCESS_KEY"`
	UseSSL          bool   `env:"S3_USE_SSL"`

	HealthCheckBucketName string        `env:"S3_HEALTH_CHECK_BUCKET_NAME" envDefault:"health-check"`
	HealthCheckObjectName string        `env:"S3_HEALTH_CHECK_OBJECT_NAME" envDefault:"latest.log"`
	HealthCheckResultTTL  time.Duration `env:"S3_HEALTH_CHECK_RESULT_TTL" envDefault:"10s"`
}

var cfg config

func RegisterHealthchecks() {
	registerHealthchecks()
}

// deprecated consider using DefaultClientFromEnv
func Client() (*minio.Client, error) {
	return DefaultClientFromEnv()
}

// Client with environment based configuration. Registers healthchecks automatically. If yo do not want to use healthchecks
// consider calling CustomClient.
func DefaultClientFromEnv() (*minio.Client, error) {
	registerHealthchecks()
	return CustomClient(cfg.Endpoint, &minio.Options{
		Secure:       cfg.UseSSL,
		Region:       cfg.Region,
		BucketLookup: minio.BucketLookupAuto,
		Creds:        credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
	})
}

// CustomClient with customized client
func CustomClient(endpoint string, opts *minio.Options) (*minio.Client, error) {
	client, err := minio.NewWithOptions(endpoint, opts)
	if err != nil {
		return nil, err
	}
	log.Logger().Info().Str("endpoint", endpoint).
		Str("region", opts.Region).
		Bool("ssl", opts.Secure).
		Msg("S3 connection created")
	client.SetCustomTransport(newCustomTransport(endpoint))
	return client, nil
}

func client() (*minio.Client, error) {
	return CustomClient(cfg.Endpoint, &minio.Options{
		Secure:       cfg.UseSSL,
		Region:       cfg.Region,
		BucketLookup: minio.BucketLookupAuto,
		Creds:        credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
	})
}

var register = &sync.Once{}

func registerHealthchecks() {
	register.Do(func() {
		// parse log config
		err := env.Parse(&cfg)
		if err != nil {
			log.Fatalf("Failed to parse object storage environment: %v", err)
		}

		client, err := client()
		if err != nil {
			log.Fatalf("Failed to create object storage client: %v", err)
		}
		servicehealthcheck.RegisterHealthCheck("objstore", &HealthCheck{
			Client: client,
		})

		ok, err := client.BucketExists(cfg.HealthCheckBucketName)
		if err != nil {
			log.Warnf("Failed to create check for bucket: %v", err)
		}
		if !ok {
			err := client.MakeBucket(cfg.HealthCheckBucketName, cfg.Region)
			if err != nil {
				log.Warnf("Failed to create bucket: %v", err)
			}
		}
	})
}

func newCustomTransport(endpoint string) http.RoundTripper {
	return transport.NewDefaultTransportChain().Use(newMetricRoundTripper(endpoint))
}
