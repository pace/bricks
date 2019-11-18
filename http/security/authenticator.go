// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/31 by Charlotte Pröller

package security

import (
	"context"
	"net/http"
)

// Authenticator describes the needed functions for authentication,
// already implemented in oauth2.Authenticator and apikey.Authenticator
type Authenticator interface {
	// Authorize should authorize a requests. This method should directly add
	// errors to the response and return a context with information of the authorization
	// and true if the authorization was successful and false if any error occurred
	Authorize(r *http.Request, w http.ResponseWriter) (context.Context, bool)
}
