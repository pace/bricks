// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/12 by Vincent Landgraf

// Package redis helps creating redis connection pools
package redis

import (
	"context"
	"time"

	"github.com/caarlos0/env"
	"github.com/go-redis/redis/v7"
	"github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
	"github.com/pace/bricks/maintenance/log"
	"github.com/prometheus/client_golang/prometheus"
)

type config struct {
	Addrs              []string      `env:"REDIS_HOSTS" envSeparator:"," envDefault:"redis:6379"`
	Password           string        `env:"REDIS_PASSWORD"`
	DB                 int           `env:"REDIS_DB"`
	MaxRetries         int           `env:"REDIS_MAX_RETRIES"`
	MinRetryBackoff    time.Duration `env:"REDIS_MIN_RETRY_BACKOFF"`
	MaxRetryBackoff    time.Duration `env:"REDIS_MAX_RETRY_BACKOFF"`
	DialTimeout        time.Duration `env:"REDIS_DIAL_TIMEOUT"`
	ReadTimeout        time.Duration `env:"REDIS_READ_TIMEOUT"`
	WriteTimeout       time.Duration `env:"REDIS_WRITE_TIMEOUT"`
	PoolSize           int           `env:"REDIS_POOL_SIZE"`
	MinIdleConns       int           `env:"REDIS_MIN_IDLE_CONNS"`
	MaxConnAge         time.Duration `env:"REDIS_MAX_CONNAGE"`
	PoolTimeout        time.Duration `env:"REDIS_POOL_TIMEOUT"`
	IdleTimeout        time.Duration `env:"REDIS_IDLE_TIMEOUT"`
	IdleCheckFrequency time.Duration `env:"REDIS_IDLE_CHECK_FREQUENCY"`
	// Name of the key that is written to check, if redis is healthy
	HealthCheckKey string `env:"REDIS_HEALTH_CHECK_KEY" envDefault:"healthy"`
	// Amount of time to cache the last health check result
	HealthCheckResultTTL time.Duration `env:"REDIS_HEALTH_CHECK_RESULT_TTL" envDefault:"10s"`
}

var (
	paceRedisCmdTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pace_redis_cmd_total",
			Help: "Collects stats about the number of redis requests made",
		},
		[]string{"method"},
	)
	paceRedisCmdFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pace_redis_cmd_failed",
			Help: "Collects stats about the number of redis requests failed",
		},
		[]string{"method"},
	)
	paceRedisCmdDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pace_redis_cmd_duration_seconds",
			Help:    "Collect performance metrics for each method",
			Buckets: []float64{.1, .25, .5, 1, 2.5, 5, 10, 60},
		},
		[]string{"method"},
	)
)

var cfg config

func init() {
	prometheus.MustRegister(paceRedisCmdTotal)
	prometheus.MustRegister(paceRedisCmdFailed)
	prometheus.MustRegister(paceRedisCmdDurationSeconds)

	// parse log config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Failed to parse redis environment: %v", err)
	}

	servicehealthcheck.RegisterHealthCheck("redis", &HealthCheck{
		Client: Client(),
	})
}

// Client with environment based configuration
func Client(overwriteOpts ...func(*redis.Options)) *redis.Client {
	opts := &redis.Options{
		Addr:               cfg.Addrs[0],
		Password:           cfg.Password,
		DB:                 cfg.DB,
		MaxRetries:         cfg.MaxRetries,
		MinRetryBackoff:    cfg.MinRetryBackoff,
		MaxRetryBackoff:    cfg.MaxRetryBackoff,
		DialTimeout:        cfg.DialTimeout,
		ReadTimeout:        cfg.ReadTimeout,
		WriteTimeout:       cfg.WriteTimeout,
		PoolSize:           cfg.PoolSize,
		MinIdleConns:       cfg.MinIdleConns,
		MaxConnAge:         cfg.MaxConnAge,
		PoolTimeout:        cfg.PoolTimeout,
		IdleTimeout:        cfg.IdleTimeout,
		IdleCheckFrequency: cfg.IdleCheckFrequency,
	}

	for _, o := range overwriteOpts {
		o(opts)
	}

	return CustomClient(opts)
}

// CustomClient with passed configuration
func CustomClient(opts *redis.Options) *redis.Client {
	log.Logger().Info().Str("addr", opts.Addr).
		Msg("Redis connection pool created")
	return redis.NewClient(opts)
}

// ClusterClient with environment based configuration
func ClusterClient() *redis.ClusterClient {
	return CustomClusterClient(&redis.ClusterOptions{
		Addrs:              cfg.Addrs,
		Password:           cfg.Password,
		MaxRetries:         cfg.MaxRetries,
		MinRetryBackoff:    cfg.MinRetryBackoff,
		MaxRetryBackoff:    cfg.MaxRetryBackoff,
		DialTimeout:        cfg.DialTimeout,
		ReadTimeout:        cfg.ReadTimeout,
		WriteTimeout:       cfg.WriteTimeout,
		PoolSize:           cfg.PoolSize,
		MinIdleConns:       cfg.MinIdleConns,
		MaxConnAge:         cfg.MaxConnAge,
		PoolTimeout:        cfg.PoolTimeout,
		IdleTimeout:        cfg.IdleTimeout,
		IdleCheckFrequency: cfg.IdleCheckFrequency,
	})
}

// CustomClusterClient with passed configuration
func CustomClusterClient(opts *redis.ClusterOptions) *redis.ClusterClient {
	log.Logger().Info().Strs("addrs", opts.Addrs).
		Msg("Redis cluster connection pool created")
	return redis.NewClusterClient(opts)
}

// WithContext adds a logging and tracing wrapper to the passed client
func WithContext(ctx context.Context, c *redis.Client) *redis.Client {
	c = c.WithContext(ctx)
	c.AddHook(&logtracer{})
	return c
}

// WithClusterContext adds a logging and tracing wrapper to the passed client
func WithClusterContext(ctx context.Context, c *redis.ClusterClient) *redis.ClusterClient {
	c = c.WithContext(ctx)
	c.AddHook(&logtracer{})
	return c
}

type logtracer struct{}

type logtracerKey struct{}

type logtracerValues struct {
	startedAt time.Time
	span      opentracing.Span
}

func (lt *logtracer) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	startedAt := time.Now()

	span, _ := opentracing.StartSpanFromContext(ctx, cmd.Name())
	span.LogFields(olog.String("cmd", cmd.Name()))
	defer span.Finish()

	span.SetTag("db.system", "redis")

	paceRedisCmdTotal.With(prometheus.Labels{
		"method": cmd.Name(),
	}).Inc()

	return context.WithValue(ctx, logtracerKey{}, &logtracerValues{
		startedAt: startedAt,
		span:      span,
	}), nil
}

func (l *logtracer) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	vals := ctx.Value(logtracerKey{}).(*logtracerValues)
	le := log.Ctx(ctx).Debug().Str("cmd", cmd.Name()).Str("sentry:category", "redis")

	// add error
	cmdErr := cmd.Err()
	if cmdErr != nil {
		vals.span.LogFields(olog.Error(cmdErr))
		le = le.Err(cmdErr)
		paceRedisCmdFailed.With(prometheus.Labels{
			"method": cmd.Name(),
		}).Inc()
	}

	// do log statement
	dur := float64(time.Since(vals.startedAt)) / float64(time.Millisecond)
	le.Float64("duration", dur).Msg("Redis query")

	paceRedisCmdDurationSeconds.With(prometheus.Labels{
		"method": cmd.Name(),
	}).Observe(dur)

	return nil
}

func (l *logtracer) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (l *logtracer) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	return nil
}
