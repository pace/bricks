// Copyright © 2022 by PACE Telematics GmbH. All rights reserved.

package postgres

import (
	"time"
)

type ConfigOption func(cfg *Config)

func WithQueryLogging(logRead, logWrite bool) ConfigOption {
	return func(cfg *Config) {
		cfg.LogRead = logRead
		cfg.LogWrite = logWrite
	}
}

// WithPort - customize the db port
func WithPort(port int) ConfigOption {
	return func(cfg *Config) {
		cfg.Port = port
	}
}

// WithHost - customise the db host
func WithHost(host string) ConfigOption {
	return func(cfg *Config) {
		cfg.Host = host
	}
}

// WithPassword - customise the db password
func WithPassword(password string) ConfigOption {
	return func(cfg *Config) {
		cfg.Password = password
	}
}

// WithUser - customise the db user
func WithUser(user string) ConfigOption {
	return func(cfg *Config) {
		cfg.User = user
	}
}

// WithDatabase - customise the db name
func WithDatabase(database string) ConfigOption {
	return func(cfg *Config) {
		cfg.Database = database
	}
}

// WithApplicationName -ApplicationName is the application name. Used in logs on Pg side.
// Only available from pg-9.0.
func WithApplicationName(applicationName string) ConfigOption {
	return func(cfg *Config) {
		cfg.ApplicationName = applicationName
	}
}

// WithMaxRetries - Maximum number of retries before giving up.
func WithMaxRetries(maxRetries int) ConfigOption {
	return func(cfg *Config) {
		cfg.MaxRetries = maxRetries
	}
}

// WithRetryStatementTimeout - Whether to retry queries cancelled because of statement_timeout.
func WithRetryStatementTimeout(retryStatementTimeout bool) ConfigOption {
	return func(cfg *Config) {
		cfg.RetryStatementTimeout = retryStatementTimeout
	}
}

// WithMinRetryBackoff - Minimum backoff between each retry.
// -1 disables backoff.
func WithMinRetryBackoff(minRetryBackoff time.Duration) ConfigOption {
	return func(cfg *Config) {
		cfg.MinRetryBackoff = minRetryBackoff
	}
}

// WithMaxRetryBackoff - Maximum backoff between each retry.
// -1 disables backoff.
func WithMaxRetryBackoff(maxRetryBackoff time.Duration) ConfigOption {
	return func(cfg *Config) {
		cfg.MaxRetryBackoff = maxRetryBackoff
	}
}

// WithDialTimeout - Dial timeout for establishing new connections.
func WithDialTimeout(dialTimeout time.Duration) ConfigOption {
	return func(cfg *Config) {
		cfg.DialTimeout = dialTimeout
	}
}

// WithReadTimeout - Timeout for socket reads. If reached, commands will fail
// with a timeout instead of blocking.
func WithReadTimeout(readTimeout time.Duration) ConfigOption {
	return func(cfg *Config) {
		cfg.ReadTimeout = readTimeout
	}
}

// WithWriteTimeout - Timeout for socket writes. If reached, commands will fail
// with a timeout instead of blocking.
func WithWriteTimeout(writeTimeout time.Duration) ConfigOption {
	return func(cfg *Config) {
		cfg.WriteTimeout = writeTimeout
	}
}

// WithPoolSize - Maximum number of socket connections.
func WithPoolSize(poolSize int) ConfigOption {
	return func(cfg *Config) {
		cfg.PoolSize = poolSize
	}
}

// WithMinIdleConns - Minimum number of idle connections which is useful when establishing
// new connection is slow.
func WithMinIdleConns(minIdleConns int) ConfigOption {
	return func(cfg *Config) {
		cfg.MinIdleConns = minIdleConns
	}
}

// WithMaxConnAge - Connection age at which client retires (closes) the connection.
// It is useful with proxies like PgBouncer and HAProxy.
func WithMaxConnAge(maxConnAge time.Duration) ConfigOption {
	return func(cfg *Config) {
		cfg.MaxConnAge = maxConnAge
	}
}

// WithPoolTimeout - Time for which client waits for free connection if all
// connections are busy before returning an error.
func WithPoolTimeout(poolTimeout time.Duration) ConfigOption {
	return func(cfg *Config) {
		cfg.PoolTimeout = poolTimeout
	}
}

// WithIdleTimeout - Amount of time after which client closes idle connections.
// Should be less than server's timeout.
// -1 disables idle timeout check.
func WithIdleTimeout(idleTimeout time.Duration) ConfigOption {
	return func(cfg *Config) {
		cfg.IdleTimeout = idleTimeout
	}
}

// WithIdleCheckFrequency - Frequency of idle checks made by idle connection's reaper.
// -1 disables idle connection's reaper,
// but idle connections are still discarded by the client
// if IdleTimeout is set.
func WithIdleCheckFrequency(idleCheckFrequency time.Duration) ConfigOption {
	return func(cfg *Config) {
		cfg.IdleCheckFrequency = idleCheckFrequency
	}
}

// WithHealthCheckTableName - Name of the Table that is created to try if database is writeable
func WithHealthCheckTableName(healthCheckTableName string) ConfigOption {
	return func(cfg *Config) {
		cfg.HealthCheckTableName = healthCheckTableName
	}
}

// WithHealthCheckResultTTL - Amount of time to cache the last health check result
func WithHealthCheckResultTTL(healthCheckResultTTL time.Duration) ConfigOption {
	return func(cfg *Config) {
		cfg.HealthCheckResultTTL = healthCheckResultTTL
	}
}
