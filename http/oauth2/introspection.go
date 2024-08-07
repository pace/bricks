// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package oauth2

import (
	"context"
	"errors"
)

// TokenIntrospecter needs to be implemented for token lookup
type TokenIntrospecter interface {
	IntrospectToken(ctx context.Context, token string) (*IntrospectResponse, error)
}

// ErrInvalidToken in case the token is not valid or expired
var ErrInvalidToken = errors.New("user token is invalid")

// ErrUpstreamConnection connection issue
var ErrUpstreamConnection = errors.New("problem connecting to the introspection endpoint")

// ErrBadUpstreamResponse the response from the server has the wrong format
var ErrBadUpstreamResponse = errors.New("bad upstream response when introspecting token")

// IntrospectResponse in case of a successful check of the
// oauth2 request
type IntrospectResponse struct {
	Active   bool   `json:"active"`
	Scope    string `json:"scope"`
	ClientID string `json:"client_id"`
	UserID   string `json:"user_id"`
	AuthTime int64  `json:"auth_time"`

	// Backend identifies the backend used for introspection. This attribute
	// exists as a convenience if you have more than one authorization backend
	// and need to distinguish between those.
	Backend interface{} `json:"-"`
}
