// Copyright © 2024 by PACE Telematics GmbH. All rights reserved.

package postgres

import "time"

type ConfigOption func(cfg *Config)

func WithQueryLogging(logRead, logWrite bool) ConfigOption {
	return func(cfg *Config) {
		cfg.LogRead = logRead
		cfg.LogWrite = logWrite
	}
}

// WithPort - customize the db port.
func WithPort(port int) ConfigOption {
	return func(cfg *Config) {
		cfg.Port = port
	}
}

// WithHost - customise the db host.
func WithHost(host string) ConfigOption {
	return func(cfg *Config) {
		cfg.Host = host
	}
}

// WithPassword - customise the db password.
func WithPassword(password string) ConfigOption {
	return func(cfg *Config) {
		cfg.Password = password
	}
}

// WithUser - customise the db user.
func WithUser(user string) ConfigOption {
	return func(cfg *Config) {
		cfg.User = user
	}
}

// WithDatabase - customise the db name.
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
