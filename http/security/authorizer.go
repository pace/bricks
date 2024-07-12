// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.

package security

import (
	"context"
	"net/http"
)

// Authorizer describes the needed functions for authorization,
// already implemented in oauth2.Authorizer and apikey.Authorizer
type Authorizer interface {
	// Authorize should authorize a request.
	// Success: returns a context with information of the authorization
	// Error: Handles Errors (logging and creating a response with the error),
	// 		  returns the unchanged context of the request and false
	Authorize(r *http.Request, w http.ResponseWriter) (context.Context, bool)
}

// CanAuthorize offers a method to check if an
// authorizer can authorize a request
type CanAuthorize interface {
	// CanAuthorizeRequest should check if a request contains the needed information to be authorized
	CanAuthorizeRequest(r http.Request) bool
}
