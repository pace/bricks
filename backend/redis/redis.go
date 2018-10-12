// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/12 by Vincent Landgraf

// Package redis helps creating redis connection pools
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/caarlos0/env"
	"github.com/go-redis/redis"
	opentracing "github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
	"github.com/rs/zerolog"
	"lab.jamit.de/pace/go-microservice/maintenance/log"
)

type config struct {
	Addrs              []string      `env:"REDIS_HOSTS" envSeparator:"," envDefault:"localhost:6379"`
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
}

var cfg config

func init() {
	// parse log config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Failed to parse redis environment: %v", err)
	}
}

// Client with environment based configuration
func Client() *redis.Client {
	return CustomClient(&redis.Options{
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
	})
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
	c.WrapProcess((&logtracer{ctx}).handle)
	return c
}

// WithClusterContext adds a logging and tracing wrapper to the passed client
func WithClusterContext(ctx context.Context, c *redis.ClusterClient) *redis.ClusterClient {
	c = c.WithContext(ctx)
	c.WrapProcess((&logtracer{ctx}).handle)
	return c
}

type logtracer struct {
	ctx context.Context
}

func (lt *logtracer) handle(realProcess func(redis.Cmder) error) func(redis.Cmder) error {
	return func(cmder redis.Cmder) error {
		// check if log context is given
		var logger *zerolog.Logger
		if lt.ctx != nil {
			logger = log.Ctx(lt.ctx)
		} else {
			logger = log.Logger()
		}

		// logging prep and tracing
		le := logger.Debug().Str("cmd", cmder.Name())
		startTime := time.Now()
		span, _ := opentracing.StartSpanFromContext(lt.ctx,
			fmt.Sprintf("Redis: %s", cmder.Name()))
		span.LogFields(olog.String("cmd", cmder.Name()))
		defer span.Finish()

		// execute redis command
		err := realProcess(cmder)

		// add error
		if err != nil {
			span.LogFields(olog.Error(err))
			le = le.Err(err)
		}

		// do log statement
		dur := float64(time.Since(startTime)) / float64(time.Millisecond)
		le.Float64("duration", dur).Msg("Redis query")

		return err
	}
}
