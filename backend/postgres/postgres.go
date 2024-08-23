// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

// Package postgres helps creating PostgreSQL connection pools
package postgres

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
	"github.com/rs/zerolog"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/go-pg/pg"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
	"github.com/pace/bricks/maintenance/log"
)

type Config struct {
	Port     int    `env:"POSTGRES_PORT" envDefault:"5432"`
	Host     string `env:"POSTGRES_HOST" envDefault:"postgres"`
	Password string `env:"POSTGRES_PASSWORD" envDefault:"mysecretpassword"`
	User     string `env:"POSTGRES_USER" envDefault:"postgres"`
	Database string `env:"POSTGRES_DB" envDefault:"postgres"`

	// ApplicationName is the application name. Used in logs on Pg side.
	// Only availaible from pg-9.0.
	ApplicationName string `env:"POSTGRES_APPLICATION_NAME" envDefault:"-"`
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
	HealthCheckTableName string `env:"POSTGRES_HEALTH_CHECK_TABLE_NAME" envDefault:"healthcheck"`
	// Amount of time to cache the last health check result
	HealthCheckResultTTL time.Duration `env:"POSTGRES_HEALTH_CHECK_RESULT_TTL" envDefault:"10s"`
	// Indicator whether write (insert,update,delete) queries should be logged
	LogWrite bool `env:"POSTGRES_LOG_WRITES" envDefault:"true"`
	// Indicator whether read (select) queries should be logged
	LogRead bool `env:"POSTGRES_LOG_READS" envDefault:"false"`
}

var (
	metricQueryTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pace_postgres_query_total",
			Help: "Collects stats about the number of postgres queries made",
		},
		[]string{"database"},
	)
	metricQueryFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pace_postgres_query_failed",
			Help: "Collects stats about the number of postgres queries failed",
		},
		[]string{"database"},
	)
	metricQueryDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pace_postgres_query_duration_seconds",
			Help:    "Collect performance metrics for each postgres query",
			Buckets: []float64{.1, .25, .5, 1, 2.5, 5, 10, 60},
		},
		[]string{"database"},
	)
	metricQueryRowsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pace_postgres_query_rows_total",
			Help: "Collects stats about the number of rows returned by a postgres query",
		},
		[]string{"database"},
	)
	metricQueryAffectedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pace_postgres_query_affected_total",
			Help: "Collects stats about the number of rows affected by a postgres query",
		},
		[]string{"database"},
	)
)

var cfg Config

func init() {
	prometheus.MustRegister(metricQueryTotal)
	prometheus.MustRegister(metricQueryFailed)
	prometheus.MustRegister(metricQueryDurationSeconds)
	prometheus.MustRegister(metricQueryRowsTotal)
	prometheus.MustRegister(metricQueryAffectedTotal)

	// parse log Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Failed to parse postgres environment: %v", err)
	}

	// if the application name is unset infer it from the:
	// jaeger service name , service name or executable name
	if cfg.ApplicationName == "-" {
		names := []string{
			os.Getenv("JAEGER_SERVICE_NAME"),
			os.Getenv("SERVICE_NAME"),
			filepath.Base(os.Args[0]),
		}
		for _, name := range names {
			if name != "" {
				cfg.ApplicationName = name
				break
			}
		}
	}

	servicehealthcheck.RegisterHealthCheck("postgresdefault", &HealthCheck{
		Pool: &pgPoolAdapter{db: DefaultConnectionPool()},
	})
}

var (
	defaultPool     *pg.DB
	defaultPoolOnce sync.Once
)

// DefaultConnectionPool returns a the default database connection pool that is
// configured using the POSTGRES_* env vars and instrumented with tracing,
// logging and metrics.
func DefaultConnectionPool() *pg.DB {
	var err error
	defaultPoolOnce.Do(func() {
		if defaultPool == nil {
			defaultPool = ConnectionPool()
			// add metrics
			metrics := NewConnectionPoolMetrics()
			prometheus.MustRegister(metrics)
			err = metrics.ObserveRegularly(context.Background(), defaultPool, "default")
		}
	})
	if err != nil {
		panic(err)
	}
	return defaultPool
}

// ConnectionPool returns a new database connection pool
// that is already configured with the correct credentials and
// instrumented with tracing and logging
// Used Config is taken from the env and it's default values. These
// values can be overwritten by the use of ConfigOption.
func ConnectionPool(opts ...ConfigOption) *pg.DB {

	// apply functional options if given to overwrite the default config / env config
	for _, f := range opts {
		f(&cfg)
	}

	return CustomConnectionPool(&pg.Options{
		Addr:                  fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		User:                  cfg.User,
		Password:              cfg.Password,
		Database:              cfg.Database,
		ApplicationName:       cfg.ApplicationName,
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

// CustomConnectionPool returns a new database connection pool
// that is already configured with the correct credentials and
// instrumented with tracing and logging using the passed options
//
// Fot a health check for this connection a PgHealthCheck needs to
// be registered:
//
//	servicehealthcheck.RegisterHealthCheck(...)
func CustomConnectionPool(opts *pg.Options) *pg.DB {
	log.Logger().Info().Str("addr", opts.Addr).
		Str("user", opts.User).
		Str("database", opts.Database).
		Str("as", opts.ApplicationName).
		Msg("PostgreSQL connection pool created")
	db := pg.Connect(opts)
	if cfg.LogWrite || cfg.LogRead {
		db.OnQueryProcessed(queryLogger)
	} else {
		log.Logger().Warn().Msg("Connection pool has logging queries disabled completely")
	}

	db.OnQueryProcessed(openTracingAdapter)
	db.OnQueryProcessed(func(event *pg.QueryProcessedEvent) {
		metricsAdapter(event, opts)
	})

	return db
}

type queryMode int

const (
	readMode  queryMode = iota
	writeMode queryMode = iota
)

// determineQueryMode is a poorman's attempt at checking whether the query is a read or write to the database.
// Feel free to improve.
func determineQueryMode(qry string) queryMode {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(qry)), "select") {
		return readMode
	}
	return writeMode
}

func queryLogger(event *pg.QueryProcessedEvent) {
	q, qe := event.UnformattedQuery()
	if qe == nil {
		if !(cfg.LogRead || cfg.LogWrite) {
			return
		}
		// we can only and should only perfom the following check if we have the information availaible
		mode := determineQueryMode(q)
		if mode == readMode && !cfg.LogRead {
			return
		}
		if mode == writeMode && !cfg.LogWrite {
			return
		}

	}
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
		Float64("duration", dur).
		Str("sentry:category", "postgres")

	// add error or result set info
	if event.Error != nil {
		le = le.Err(event.Error)
	} else {
		le = le.Int("affected", event.Result.RowsAffected()).
			Int("rows", event.Result.RowsReturned())
	}

	if qe != nil {
		// this is only a display issue not a "real" issue
		le.Msgf("%v", qe)
	}
	le.Msg(q)
}

var reQueryType = regexp.MustCompile(`(\s)`)
var reQueryTypeCleanup = regexp.MustCompile(`(?m)(\s+|\n)`)

func getQueryType(s string) string {
	s = reQueryTypeCleanup.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)

	p := reQueryType.FindStringIndex(s)
	if len(p) > 0 {
		return strings.ToUpper(s[:p[0]])
	}
	return strings.ToUpper(s)
}

func openTracingAdapter(event *pg.QueryProcessedEvent) {
	// start span with general info
	q, qe := event.UnformattedQuery()
	if qe != nil {
		// this is only a display issue not a "real" issue
		q = qe.Error()
	}

	span, _ := opentracing.StartSpanFromContext(event.DB.Context(), "sql: "+getQueryType(q),
		opentracing.StartTime(event.StartTime))

	span.SetTag("db.system", "postgres")

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

	metricQueryTotal.With(labels).Inc()

	if event.Error != nil {
		metricQueryFailed.With(labels).Inc()
	} else {
		r := event.Result
		metricQueryRowsTotal.With(labels).Add(float64(r.RowsReturned()))
		metricQueryAffectedTotal.With(labels).Add(math.Max(0, float64(r.RowsAffected())))
	}
	metricQueryDurationSeconds.With(labels).Observe(dur)
}
