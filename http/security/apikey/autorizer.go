// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/30 by Charlotte Pröller

package apikey

import (
	"context"
	"net/http"

	"github.com/caarlos0/env"
	"github.com/pace/bricks/http/security"
	"github.com/pace/bricks/maintenance/log"
)

type config struct {
	ApiKey string `env:"SECURITY_API_KEY"`
}

type Authorizer struct {
}

var cfg config

// AuthorizerConfig is the interface for the Information of the Security Schema information from the API definition that we use.
// Currently we suspect that we have a apikey in the header that starts with "bearing "
type AuthorizerConfig interface {
	// getName returns the Name of the field in the header
	getName() string
}

func init() {
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Failed to parse api key environment: %v", err)
	}
}

func (a *Authorizer) Authorize(aCfg AuthorizerConfig, r *http.Request, w http.ResponseWriter) (context.Context, bool) {
	key := security.GetBearerTokenFromHeader(r, w, aCfg.getName())
	if key == "" {
		return nil, false
	}
	if key == cfg.ApiKey {
		return security.WithBearerToken(r.Context(), key), true
	}
	return nil, false
}
