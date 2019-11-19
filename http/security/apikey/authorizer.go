// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/30 by Charlotte Pröller

package apikey

import (
	"context"
	"errors"
	"net/http"

	"github.com/pace/bricks/http/security"
	"github.com/pace/bricks/maintenance/log"
)

// Authorizer implements the security.Authorizer Interface for a ApiKey based
type Authorizer struct {
	authConfig *Config
	apiKey     string
}

// Config is the Struct for the Information of the Security Schema information from the API definition that we use.
// Currently we suspect that we have an apikey in the header that is prepended by "Bearer "
type Config struct {
	Description string
	In          string
	Name        string
}

type bearerToken struct {
	value string
}

// GetValue returns the be
func (b *bearerToken) GetValue() string {
	return b.value
}

// NewAuthenticator Returns New Authorizer With the Config from the APi definition and a ApiKey, that is matched.
func NewAuthenticator(authConfig *Config, apiKey string) *Authorizer {
	return &Authorizer{authConfig: authConfig, apiKey: apiKey}
}

// Authorize authorizes a request based on the Authorization Configuration of the environment and of the config of the
// security schema.
// Success: A Context with the Token and true
// Error: the unchanged request context and false. All errors are directly written to the response
func (a *Authorizer) Authorize(r *http.Request, w http.ResponseWriter) (context.Context, bool) {
	key, err := security.GetBearerTokenFromHeader(r, a.authConfig.Name)
	if err != nil {
		log.Req(r).Info().Msg(err.Error())
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return r.Context(), false
	}
	if key == a.apiKey {
		return security.ContextWithTokenKey(r.Context(), &bearerToken{key}), true
	}
	err = errors.New("ApiKey not valid")
	log.Req(r).Info().Msg(err.Error())
	http.Error(w, err.Error(), http.StatusUnauthorized)
	return r.Context(), false
}
