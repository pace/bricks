// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/30 by Charlotte Pröller

package oauth2

import (
	"context"
	"net/http"
)

type Authorizer struct {
}

// AuthorizerConfig is the interface for the Information of the Security Schema information from the API definition that we use.
// currently this is nothing
type AuthorizerConfig interface {
}

// Authorize authorizes a request
// does a introspection and validates the scope
// success: context with introspection informations, isOk = true
// error: writes all errors directly to response, isOk = false

func (a *Authorizer) Authorize(cfg AuthorizerConfig, introspector TokenIntrospector, scope string, r *http.Request, w http.ResponseWriter) (context.Context, bool) {
	ctx, isOk := introspectRequest(r, w, introspector)
	// Check if introspection was successful
	if !isOk {
		return ctx, isOk
	}
	isOk = validateScope(ctx, w, "pay:payment-methods:create")
	// Check if the scope is valid for this user
	return ctx, isOk
}
