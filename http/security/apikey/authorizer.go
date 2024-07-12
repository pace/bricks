// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.

package apikey

import (
	"context"
	"net/http"

	"github.com/pace/bricks/http/security"
	"github.com/pace/bricks/maintenance/log"
)

// Authorizer implements the security.Authorizer interface for an api key based authorization.
type Authorizer struct {
	authConfig *Config
	apiKey     string
}

// Config contains the configuration of the security schema with type "apiKey".
type Config struct {
	// currently not used
	Description string
	// Must be "Header"
	In string
	// Header field name, should never be "Authorization" if OAuth2 and ApiKey Authorization is combined
	Name string
}

type token struct {
	value string
}

// GetValue returns the api key
func (b *token) GetValue() string {
	return b.value
}

// NewAuthorizer returns a new Authorizer for api key authorization with a config and a valid api key.
func NewAuthorizer(authConfig *Config, apiKey string) *Authorizer {
	return &Authorizer{authConfig: authConfig, apiKey: apiKey}
}

// Authorize authorizes a request based on the configured api key the config of the security schema
// Success: A context with a token containing the api key and true
// Error: the unchanged request context and false. the response already contains the error message
func (a *Authorizer) Authorize(r *http.Request, w http.ResponseWriter) (context.Context, bool) {
	key := security.GetBearerTokenFromHeader(r.Header.Get(a.authConfig.Name))
	if key == "" {
		log.Req(r).Info().Msg("No Api Key present in field " + a.authConfig.Name)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return r.Context(), false
	}
	if key == a.apiKey {
		return security.ContextWithToken(r.Context(), &token{key}), true
	}
	http.Error(w, "ApiKey not valid", http.StatusUnauthorized)
	return r.Context(), false
}

// CanAuthorizeRequest returns true, if the request contains a token in the configured header, otherwise false
func (a *Authorizer) CanAuthorizeRequest(r http.Request) bool {
	return security.GetBearerTokenFromHeader(r.Header.Get(a.authConfig.Name)) != ""
}
