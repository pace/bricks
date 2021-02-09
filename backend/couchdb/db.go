// Copyright © 2021 by PACE Telematics GmbH. All rights reserved.
// Created at 2021/02/08 by Vincent Landgraf

package couchdb

import (
	"context"

	"github.com/caarlos0/env"
	"github.com/go-kivik/couchdb/v3"
	kivik "github.com/go-kivik/kivik/v3"
	"github.com/pace/bricks/http/transport"
	"github.com/pace/bricks/maintenance/health/servicehealthcheck"
	"github.com/pace/bricks/maintenance/log"
)

func DefaultDatabase() (*kivik.DB, error) {
	return Database("")
}

func Database(name string) (*kivik.DB, error) {
	ctx := log.WithContext(context.Background())

	cfg, err := ParseConfig()
	if err != nil {
		return nil, err
	}
	client, err := Client(cfg)
	if err != nil {
		return nil, err
	}

	// use default db
	if name == "" {
		name = cfg.Database
	}

	db := client.DB(ctx, name, nil)
	if db.Err() != nil {
		return nil, db.Err()
	}

	servicehealthcheck.RegisterHealthCheck("couchdb("+name+")", &HealthCheck{
		Name:   name,
		Client: client,
		DB:     db,
		Config: cfg,
	})

	return db, nil
}

func Client(cfg *Config) (*kivik.Client, error) {
	ctx := log.WithContext(context.Background())

	client, err := kivik.New("couch", cfg.URL)
	if err != nil {
		return nil, err
	}

	chain := transport.Chain(
		&AuthTransport{
			Username: cfg.User,
			Password: cfg.Password,
		},
		&transport.JaegerRoundTripper{},
		transport.NewDumpRoundTripperEnv(),
		&transport.LoggingRoundTripper{},
	)
	tr := couchdb.SetTransport(chain)
	err = client.Authenticate(ctx, tr)
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
