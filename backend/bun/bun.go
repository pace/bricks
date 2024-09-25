package bun

import (
	"context"
	"database/sql"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/pace/bricks/backend/bun/hooks"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

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
	// Dial timeout for establishing new connections.
	DialTimeout time.Duration `env:"POSTGRES_DIAL_TIMEOUT" envDefault:"5s"`
	// Name of the Table that is created to try if database is writeable
	HealthCheckTableName string `env:"POSTGRES_HEALTH_CHECK_TABLE_NAME" envDefault:"healthcheck"`
	// Amount of time to cache the last health check result
	HealthCheckResultTTL time.Duration `env:"POSTGRES_HEALTH_CHECK_RESULT_TTL" envDefault:"10s"`
	// Indicator whether write (insert,update,delete) queries should be logged
	LogWrite bool `env:"POSTGRES_LOG_WRITES" envDefault:"true"`
	// Indicator whether read (select) queries should be logged
	LogRead bool `env:"POSTGRES_LOG_READS" envDefault:"false"`
	// Timeout for socket reads. If reached, commands will fail
	// with a timeout instead of blocking.
	ReadTimeout time.Duration `env:"POSTGRES_READ_TIMEOUT" envDefault:"30s"`
	// Timeout for socket writes. If reached, commands will fail
	// with a timeout instead of blocking.
	WriteTimeout time.Duration `env:"POSTGRES_WRITE_TIMEOUT" envDefault:"30s"`
}

var (
	cfg Config
)

func init() {
	prometheus.MustRegister(hooks.MetricQueryTotal)
	prometheus.MustRegister(hooks.MetricQueryFailed)
	prometheus.MustRegister(hooks.MetricQueryDurationSeconds)
	prometheus.MustRegister(hooks.MetricQueryAffectedTotal)

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
		db: NewDB(context.Background()),
	})
}

func NewDB(ctx context.Context, options ...ConfigOption) *bun.DB {
	for _, opt := range options {
		opt(&cfg)
	}

	connector := pgdriver.NewConnector(
		pgdriver.WithAddr(net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))),
		pgdriver.WithApplicationName(cfg.ApplicationName),
		pgdriver.WithDatabase(cfg.Database),
		pgdriver.WithDialTimeout(cfg.DialTimeout),
		pgdriver.WithPassword(cfg.Password),
		pgdriver.WithReadTimeout(cfg.ReadTimeout),
		pgdriver.WithUser(cfg.User),
		pgdriver.WithWriteTimeout(cfg.WriteTimeout),
	)

	sqldb := sql.OpenDB(connector)
	db := bun.NewDB(sqldb, pgdialect.New())

	log.Ctx(ctx).Info().Str("addr", connector.Config().Addr).
		Str("user", connector.Config().User).
		Str("database", connector.Config().Database).
		Str("as", connector.Config().AppName).
		Msg("PostgreSQL connection pool created")

	// Add hooks
	db.AddQueryHook(&hooks.TracingHook{})
	db.AddQueryHook(hooks.NewMetricsHook(cfg.Host, cfg.Database))

	if cfg.LogWrite || cfg.LogRead {
		db.AddQueryHook(hooks.NewLoggingHook(cfg.LogRead, cfg.LogWrite))
	} else {
		log.Ctx(ctx).Warn().Msg("Connection pool has logging queries disabled completely")
	}

	return db
}
