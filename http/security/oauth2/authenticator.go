// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/30 by Charlotte Pröller

package oauth2

import (
	"context"
	"net/http"
)

// Authenticator is a Implementation of security.Authenticator for oauth2
// it offers introspection and a scope that is used for authentication
type Authenticator struct {
	introspection TokenIntrospector
	scope         Scope
	config        *Config
}

// Flow is a part of the oauth2 config from the security schema
type Flow struct {
	AuthorizationURL string
	TokenURL         string
	RefreshURL       string
	Scopes           map[string]string
}

// Config contains the configuration from the api definition
type Config struct {
	Description       string
	Implicit          *Flow
	Password          *Flow
	ClientCredentials *Flow
	AuthorizationCode *Flow
}

// NewAuthenticator creates a Authenticator for a specific TokenIntrospector
// This Authenticator does not check the scope until a scope is added
func NewAuthenticator(introspector TokenIntrospector, cfg *Config) *Authenticator {
	return &Authenticator{introspection: introspector, config: cfg}
}

// WithScope returns a new Authenticator with the same TokenIntrospector and the same Config that also checks the scope of a request
func (a *Authenticator) WithScope(tok string) *Authenticator {
	return &Authenticator{introspection: a.introspection, config: a.config, scope: Scope(tok)}
}

// Authorize authorizes a request
// does a introspection and validates the scope
// checks the config even if we don't use it
// success: context with introspection information, isOk = true
// error: writes all errors directly to response, isOk = false
func (a *Authenticator) Authorize(r *http.Request, w http.ResponseWriter) (context.Context, bool) {
	ctx, ok := introspectRequest(r, w, a.introspection)
	// Check if introspection was successful
	if !ok {
		return ctx, ok
	}
	// if the Authenticator has no scope, the request is valid without looking for the scope.
	if a.scope != "" {
		// Check if the scope is valid for this user
		ok = validateScope(ctx, w, a.scope)
	}
	return ctx, ok
}
