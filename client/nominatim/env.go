// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/10/12 by Vincent Landgraf

package nominatim

import (
	"log"

	"github.com/caarlos0/env"
)

var DefaultClient *Client

type config struct {
	Endpoint string `env:"NOMINATIM_ENDPOINT" envDefault:"https://maps.pacelink.net"`
}

var cfg config

func init() {
	// parse log config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Failed to parse nominatim environment: %v", err)
	}
	DefaultClient = &Client{Endpoint: cfg.Endpoint}
}
