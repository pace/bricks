// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package http

import (
	golog "log"
	"net/http"
	"strconv"
	"time"

	"github.com/caarlos0/env/v11"

	"github.com/pace/bricks/maintenance/log"
)

func init() {
	parseConfig()
}

type config struct {
	Addr           string        `env:"ADDR"`
	Port           int           `env:"PORT" envDefault:"3000"`
	Environment    string        `env:"ENVIRONMENT" envDefault:"edge"`
	MaxHeaderBytes int           `env:"MAX_HEADER_BYTES" envDefault:"1048576"` // 1MB
	IdleTimeout    time.Duration `env:"IDLE_TIMEOUT" envDefault:"1h"`
	ReadTimeout    time.Duration `env:"READ_TIMEOUT" envDefault:"60s"`
	WriteTimeout   time.Duration `env:"WRITE_TIMEOUT" envDefault:"60s"`
}

// addrOrPort returns ADDR if it is defined, otherwise PORT is used.
func (cfg config) addrOrPort() string {
	if cfg.Addr != "" {
		return cfg.Addr
	}

	return ":" + strconv.Itoa(cfg.Port)
}

var cfg config

func parseConfig() {
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to parse server environment: %v", err)
	}
}

// Server returns a http.Server configured using environment variables,
// following https://12factor.net/.
func Server(handler http.Handler) *http.Server {
	return &http.Server{
		Addr:           cfg.addrOrPort(),
		Handler:        handler,
		ReadTimeout:    cfg.ReadTimeout,
		WriteTimeout:   cfg.WriteTimeout,
		MaxHeaderBytes: cfg.MaxHeaderBytes,
		IdleTimeout:    cfg.IdleTimeout,
		ErrorLog:       golog.New(log.Logger(), "[http.Server] ", 0),
	}
}

// Environment returns the name of the current server environment.
func Environment() string {
	return cfg.Environment
}
