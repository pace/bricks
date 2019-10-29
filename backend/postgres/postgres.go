// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/12 by Vincent Landgraf

// Package postgres helps creating PostgreSQL connection pools
package postgres

import (
	"fmt"
	"math"
	"time"

	"github.com/caarlos0/env"
	"github.com/go-pg/pg"
	"github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
	"github.com/pace/bricks/maintenance/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type config struct {
	Port     int    `env:"POSTGRES_PORT" envDefault:"5432"`
	Host     string `env:"POSTGRES_HOST" envDefault:"postgres"`
	Password string `env:"POSTGRES_PASSWORD" envDefault:"mysecretpassword"`
	User     string `env:"POSTGRES_USER" envDefault:"postgres"`
	Database string `env:"POSTGRES_DB" envDefault:"postgres"`
	// Maximum number of retries before giving up.
	MaxRetries int `env:"POSTGRES_MAX_RETRIES" envDefault:"5"`
	// Whether to retry queries cancelled because of statement_timeout.
	RetryStatementTimeout bool `env:"POSTGRES_RETRY_STATEMENT_TIMEOUT" envDefault:"false"`
	// Minimum backoff between each retry.
	// -1 disables backoff.
	MinRetryBackoff time.Duration `env:"POSTGRES_MIN_RETRY_BACKOFF" envDefault:"250ms"`
	// Maximum backoff between each retry.
	// -1 disables backoff.
	MaxRetryBackoff time.Duration `env:"POSTGRES_MAX_RETRY_BACKOFF" envDefault:"4s"`
	// Dial timeout for establishing new connections.
	DialTimeout time.Duration `env:"POSTGRES_DIAL_TIMEOUT" envDefault:"5s"`
	// Timeout for socket reads. If reached, commands will fail
	// with a timeout instead of blocking.
	ReadTimeout time.Duration `env:"POSTGRES_READ_TIMEOUT" envDefault:"30s"`
	// Timeout for socket writes. If reached, commands will fail
	// with a timeout instead of blocking.
	WriteTimeout time.Duration `env:"POSTGRES_WRITE_TIMEOUT" envDefault:"30s"`
	// Maximum number of socket connections.
	PoolSize int `env:"POSTGRES_POOL_SIZE" envDefault:"100"`
	// Minimum number of idle connections which is useful when establishing
	// new connection is slow.
	MinIdleConns int `env:"POSTGRES_MIN_IDLE_CONNECTIONS" envDefault:"10"`
	// Connection age at which client retires (closes) the connection.
	// It is useful with proxies like PgBouncer and HAProxy.
	MaxConnAge time.Duration `env:"POSTGRES_MAX_CONN_AGE" envDefault:"30m"`
	// Time for which client waits for free connection if all
	// connections are busy before returning an error.
	PoolTimeout time.Duration `env:"POSTGRES_POOL_TIMEOUT" envDefault:"31s"`
	// Amount of time after which client closes idle connections.
	// Should be less than server's timeout.
	// -1 disables idle timeout check.
	IdleTimeout time.Duration `env:"POSTGRES_IDLE_TIMEOUT" envDefault:"5m"`
	// Frequency of idle checks made by idle connections reaper.
	// -1 disables idle connections reaper,
	// but idle connections are still discarded by the client
	// if IdleTimeout is set.
	IdleCheckFrequency time.Duration `env:"POSTGRES_IDLE_CHECK_FREQUENCY" envDefault:"1m"`
	// Name of the Table that is created to try if database is writeable
	HealthTableName string `env:"POSTGRES_HEALTHCHECK_TABLE" envDefault:"healthcheck"`
	// Amount of time to cache the last health check result
	HealthMaxRequest time.Duration `env:"POSTGRES_HEALTHCHECK_MAX_REQUEST_SEC" envDefault:"10s"`
}

var (
	pacePostgresQueryTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pace_postgres_query_total",
			Help: "Collects stats about the number of postgres queries made",
		},
		[]string{"database"},
	)
	pacePostgresQueryFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pace_postgres_query_failed",
			Help: "Collects stats about the number of postgres queries failed",
		},
		[]string{"database"},
	)
	pacePostgresQueryDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pace_postgres_query_duration_seconds",
			Help:    "Collect performance metrics for each postgres query",
			Buckets: []float64{.1, .25, .5, 1, 2.5, 5, 10, 60},
		},
		[]string{"database"},
	)
	pacePostgresQueryRowsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pace_postgres_query_rows_total",
			Help: "Collects stats about the number of rows returned by a postgres query",
		},
		[]string{"database"},
	)
	pacePostgresQueryAffectedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pace_postgres_query_affected_total",
			Help: "Collects stats about the number of rows affected by a postgres query",
		},
		[]string{"database"},
	)
)

var cfg config

func init() {
	prometheus.MustRegister(pacePostgresQueryTotal)
	prometheus.MustRegister(pacePostgresQueryFailed)
	prometheus.MustRegister(pacePostgresQueryDurationSeconds)
	prometheus.MustRegister(pacePostgresQueryRowsTotal)
	prometheus.MustRegister(pacePostgresQueryAffectedTotal)

	// parse log config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Failed to parse postgres environment: %v", err)
	}
	servicehealthcheck.RegisterHealthCheck(&PgHealthCheck{}, "postgresdefault")

}

// ConnectionPool returns a new database connection Pool
// that is already configured with the correct credentials and
// instrumented with tracing and logging
func ConnectionPool() *pg.DB {
	return CustomConnectionPool(&pg.Options{
		Addr:                  fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		User:                  cfg.User,
		Password:              cfg.Password,
		Database:              cfg.Database,
		MaxRetries:            cfg.MaxRetries,
		RetryStatementTimeout: cfg.RetryStatementTimeout,
		MinRetryBackoff:       cfg.MinRetryBackoff,
		MaxRetryBackoff:       cfg.MaxRetryBackoff,
		DialTimeout:           cfg.DialTimeout,
		ReadTimeout:           cfg.ReadTimeout,
		WriteTimeout:          cfg.WriteTimeout,
		PoolSize:              cfg.PoolSize,
		MinIdleConns:          cfg.MinIdleConns,
		MaxConnAge:            cfg.MaxConnAge,
		PoolTimeout:           cfg.PoolTimeout,
		IdleTimeout:           cfg.IdleTimeout,
		IdleCheckFrequency:    cfg.IdleCheckFrequency,
	})
}

// CustomConnectionPool returns a new database connection Pool
// that is already configured with the correct credentials and
// instrumented with tracing and logging using the passed options
//
// Fot a health check for this connection a PgHealthCheck needs to
// be registered:
//  servicehealthcheck.RegisterHealthCheck(...)
func CustomConnectionPool(opts *pg.Options) *pg.DB {
	log.Logger().Info().Str("addr", opts.Addr).
		Str("user", opts.User).Str("database", opts.Database).
		Msg("PostgreSQL connection Pool created")
	db := pg.Connect(opts)
	db.OnQueryProcessed(queryLogger)
	db.OnQueryProcessed(openTracingAdapter)
	db.OnQueryProcessed(func(event *pg.QueryProcessedEvent) {
		metricsAdapter(event, opts)
	})
	return db
}

func queryLogger(event *pg.QueryProcessedEvent) {
	ctx := event.DB.Context()
	dur := float64(time.Since(event.StartTime)) / float64(time.Millisecond)

	// check if log context is given
	var logger *zerolog.Logger
	if ctx != nil {
		logger = log.Ctx(ctx)
	} else {
		logger = log.Logger()
	}

	// add general info
	le := logger.Debug().
		Str("file", event.File).
		Int("line", event.Line).
		Str("func", event.Func).
		Int("attempt", event.Attempt).
		Float64("duration", dur)

	// add error or result set info
	if event.Error != nil {
		le = le.Err(event.Error)
	} else {
		le = le.Int("affected", event.Result.RowsAffected()).
			Int("rows", event.Result.RowsReturned())
	}

	q, qe := event.UnformattedQuery()
	if qe != nil {
		// this is only a display issue not a "real" issue
		le.Msgf("%v", qe)
	}
	le.Msg(q)
}

func openTracingAdapter(event *pg.QueryProcessedEvent) {
	// start span with general info
	q, qe := event.UnformattedQuery()
	if qe != nil {
		// this is only a display issue not a "real" issue
		q = qe.Error()
	}

	name := fmt.Sprintf("PostgreSQL: %s", q)
	span, _ := opentracing.StartSpanFromContext(event.DB.Context(), name,
		opentracing.StartTime(event.StartTime))

	fields := []olog.Field{
		olog.String("file", event.File),
		olog.Int("line", event.Line),
		olog.String("func", event.Func),
		olog.Int("attempt", event.Attempt),
		olog.String("query", q),
	}

	// add error or result set info
	if event.Error != nil {
		fields = append(fields, olog.Error(event.Error))
	} else {
		fields = append(fields,
			olog.Int("affected", event.Result.RowsAffected()),
			olog.Int("rows", event.Result.RowsReturned()))
	}

	span.LogFields(fields...)
	span.Finish()
}

func metricsAdapter(event *pg.QueryProcessedEvent, opts *pg.Options) {
	dur := float64(time.Since(event.StartTime)) / float64(time.Millisecond)
	labels := prometheus.Labels{
		"database": opts.Addr + "/" + opts.Database,
	}

	pacePostgresQueryTotal.With(labels).Inc()

	if event.Error != nil {
		pacePostgresQueryFailed.With(labels).Inc()
	} else {
		r := event.Result
		pacePostgresQueryRowsTotal.With(labels).Add(float64(r.RowsReturned()))
		pacePostgresQueryAffectedTotal.With(labels).Add(math.Max(0, float64(r.RowsAffected())))
	}
	pacePostgresQueryDurationSeconds.With(labels).Observe(dur)
}
