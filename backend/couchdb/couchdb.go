// Copyright © 2021 by PACE Telematics GmbH. All rights reserved.
// Created at 2021/02/08 by Vincent Landgraf

package couchdb

import "time"

type Config struct {
	URL                  string        `env:"COUCHDB_URL" envDefault:"http://couchdb:5984/"`
	User                 string        `env:"COUCHDB_USER" envDefault:"admin"`
	Password             string        `env:"COUCHDB_PASSWORD" envDefault:"secret"`
	Database             string        `env:"COUCHDB_DB" envDefault:"test"`
	DatabaseAutoCreate   bool          `env:"COUCHDB_DB_AUTO_CREATE" envDefault:"true"`
	HealthCheckKey       string        `env:"COUCHDB_HEALTH_CHECK_KEY" envDefault:"$health_check"`
	HealthCheckResultTTL time.Duration `env:"COUCHDB_HEALTH_CHECK_RESULT_TTL" envDefault:"10s"`
	DisableHealthCheck   bool          `env:"COUCHDB_DISABLE_HEALTH_CHECK" envDefault:"false"`
}
