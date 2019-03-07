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
	opentracing "github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/pace/bricks/maintenance/log"
)

type config struct {
	Port     int    `env:"POSTGRES_PORT" envDefault:"5432"`
	Host     string `env:"POSTGRES_HOST" envDefault:"postgres"`
	Password string `env:"POSTGRES_PASSWORD" envDefault:"mysecretpassword"`
	User     string `env:"POSTGRES_USER" envDefault:"postgres"`
	Database string `env:"POSTGRES_DB" envDefault:"postgres"`
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
}

// ConnectionPool returns a new database connection pool
// that is already configured with the correct credentials and
// instrumented with tracing and logging
func ConnectionPool() *pg.DB {
	return CustomConnectionPool(&pg.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		User:     cfg.User,
		Password: cfg.Password,
		Database: cfg.Database,
	})
}

// CustomConnectionPool returns a new database connection pool
// that is already configured with the correct credentials and
// instrumented with tracing and logging using the passed options
func CustomConnectionPool(opts *pg.Options) *pg.DB {
	log.Logger().Info().Str("addr", opts.Addr).
		Str("user", opts.User).Str("database", opts.Database).
		Msg("PostgreSQL connection pool created")
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
