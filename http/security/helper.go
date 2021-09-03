// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/30 by Charlotte Pröller

package security

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"
)

// Token represents an authentication token.
type Token interface {
	// GetValue returns the bearer token
	GetValue() string
}
type ctx string

type TokenString string

func (ts TokenString) GetValue() string {
	return string(ts)
}

// prefix of the Authorization header
const headerPrefix = "Bearer "

var tokenKey = ctx("Token")

// GetBearerTokenFromHeader removes the prefix of authHeader
// authHeader should be the value of the authorization header field.
// success: returns the bearer token without the prefix
// error: return an empty token
func GetBearerTokenFromHeader(authHeader string) string {
	hasPrefix := strings.HasPrefix(authHeader, headerPrefix)
	if !hasPrefix {
		return ""
	}
	return strings.TrimPrefix(authHeader, headerPrefix)
}

// ContextWithToken creates a new Context with the token
func ContextWithToken(targetCtx context.Context, token Token) context.Context {
	if token != nil {
		targetCtx = metadata.AppendToOutgoingContext(targetCtx, "bearer_token", token.GetValue())
	}
	return context.WithValue(targetCtx, tokenKey, token)
}

// GetTokenFromContext returns the token, if it is stored in the context, and true if the token is present.
func GetTokenFromContext(ctx context.Context) (Token, bool) {
	val := ctx.Value(tokenKey)
	if val == nil {
		return nil, false
	}
	tok, ok := val.(Token)
	return tok, ok
}

// GetAuthHeader creates the valid value for the authentication header
func GetAuthHeader(tok Token) string {
	return headerPrefix + tok.GetValue()
}
