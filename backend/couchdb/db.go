// Copyright © 2021 by PACE Telematics GmbH. All rights reserved.

package couchdb

import (
	"net/http"

	"github.com/caarlos0/env/v11"
	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/couchdb"

	"github.com/pace/bricks/http/transport"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
)

func DefaultDatabase() (*kivik.DB, error) {
	return Database("")
}

func Database(name string) (*kivik.DB, error) {
	cfg, err := ParseConfig()
	if err != nil {
		return nil, err
	}
	// Primary client+db
	_, db, err := clientAndDB(name, cfg)
	if err != nil {
		return nil, err
	}

	// Secondary (healthcheck) client+db
	healthCheckClient, healthCheckDB, err := clientAndDB(name, cfg)
	if err != nil {
		return nil, err
	}

	if !cfg.DisableHealthCheck {
		servicehealthcheck.RegisterHealthCheck("couchdb("+name+")", &HealthCheck{
			Name:   name,
			Client: healthCheckClient,
			DB:     healthCheckDB,
			Config: cfg,
		})
	}

	return db, nil
}

func clientAndDB(dbName string, cfg *Config) (*kivik.Client, *kivik.DB, error) {
	client, err := Client(cfg)
	if err != nil {
		return nil, nil, err
	}

	// use default db
	if dbName == "" {
		dbName = cfg.Database
	}

	db := client.DB(dbName)
	if db.Err() != nil {
		return nil, nil, db.Err()
	}
	return client, db, err
}

func Client(cfg *Config) (*kivik.Client, error) {
	rts := []transport.ChainableRoundTripper{
		&AuthTransport{
			Username: cfg.User,
			Password: cfg.Password,
		},
		transport.NewDumpRoundTripperEnv(),
	}
	if !cfg.DisableRequestLogging {
		rts = append(rts, &transport.LoggingRoundTripper{})
	}

	client, err := kivik.New("couch", cfg.URL, couchdb.OptionHTTPClient(&http.Client{Transport: transport.Chain(rts...)}))
	if err != nil {
		return nil, err
	}

	return client, nil
}

func ParseConfig() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
