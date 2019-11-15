// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/30 by Charlotte Pröller

package apikey

import (
	"context"
	"net/http"

	"github.com/pace/bricks/http/security"
	"github.com/pace/bricks/maintenance/errors"
)

// Authenticator implements the security.Authenticator Interface for a ApiKey based
type Authenticator struct {
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

// NewAuthenticator Returns New Authenticator With the Config from the APi definition and a ApiKey, that is matched.
func NewAuthenticator(authConfig *Config, apiKey string) *Authenticator {
	return &Authenticator{authConfig: authConfig, apiKey: apiKey}
}

// Authorize authorizes a request based on the Authorization Configuration of the environment and of the config of the
// security schema.
// Success: A Context with the Token and true
// Error: the unchanged request context and false. All errors are directly written to the response
func (a *Authenticator) Authorize(r *http.Request, w http.ResponseWriter) (context.Context, bool) {
	key := security.GetBearerTokenFromHeader(r, w, a.authConfig.Name)
	if key == "" {
		return r.Context(), false
	}
	if key == a.apiKey {
		return security.WithBearerToken(r.Context(), key), true
	}
	http.Error(w, errors.New("ApiKey not valid").Error(), http.StatusUnauthorized)
	return r.Context(), false
}
