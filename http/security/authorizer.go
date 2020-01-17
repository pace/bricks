// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/31 by Charlotte Pröller

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

type NoOpWriter struct {
}

func (n NoOpWriter) Header() http.Header {
	//Noop
	return nil
}

func (n NoOpWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (n NoOpWriter) WriteHeader(statusCode int) {
	//noop
}
