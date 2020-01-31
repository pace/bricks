// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/30 by Charlotte Pröller

package oauth2

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pace/bricks/http/security"
)

// Authorizer is an implementation of security.Authorizer for OAuth2
// it uses introspection to get user data and can check the scope
type Authorizer struct {
	introspection TokenIntrospecter
	scope         Scope
	config        *Config
}

// Flow is a part of the OAuth2 config from the security schema
type Flow struct {
	AuthorizationURL string
	TokenURL         string
	RefreshURL       string
	Scopes           map[string]string
}

// Config contains the configuration from the api definition - currently not used
type Config struct {
	Description       string
	Implicit          *Flow
	Password          *Flow
	ClientCredentials *Flow
	AuthorizationCode *Flow
}

// NewAuthorizer creates an Authorizer for a specific TokenIntrospecter
// This Authorizer does not check the scope
func NewAuthorizer(introspector TokenIntrospecter, cfg *Config) *Authorizer {
	return &Authorizer{introspection: introspector, config: cfg}
}

// WithScope returns a new Authorizer with the same TokenIntrospecter and the same Config that also checks the scope of a request
func (a *Authorizer) WithScope(tok string) *Authorizer {
	return &Authorizer{introspection: a.introspection, config: a.config, scope: Scope(tok)}
}

// Authorize authorizes a request with an introspection and validates the scope
// Success: returns context with the introspection result and true
// Error: writes all errors directly to response, returns unchanged context and false
func (a *Authorizer) Authorize(r *http.Request, w http.ResponseWriter) (context.Context, bool) {
	ctx, ok := introspectRequest(r, w, a.introspection)
	// Check if introspection was successful
	if !ok {
		return ctx, ok
	}
	// If the Authorizer has no scope, the request is valid, otherwise the scope must be validated
	if a.scope != "" {
		// Check if the scope is valid for this user
		ok = validateScope(ctx, w, a.scope)
	}
	return ctx, ok
}

func validateScope(ctx context.Context, w http.ResponseWriter, req Scope) bool {
	if !HasScope(ctx, req) {
		http.Error(w, fmt.Sprintf("Forbidden - requires scope %q", req), http.StatusForbidden)
		return false
	}
	return true
}

// CanAuthorizeRequest returns true, if the request contains a token in the configured header, otherwise false
func (a *Authorizer) CanAuthorizeRequest(r *http.Request) bool {
	return security.GetBearerTokenFromHeader(r.Header.Get(oAuth2Header)) != ""
}
