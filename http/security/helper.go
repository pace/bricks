// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/30 by Charlotte Pröller

package security

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/pace/bricks/maintenance/errors"
)

// Token represents an authentication token.
type Token interface {
	// GetValue returns the bearer token
	GetValue() string
}

type ctx string

// HeaderPrefix prefix of the Authentication value in the header
const headerPrefix = "Bearer "

var tokenKey = ctx("Token")

// GetBearerTokenFromHeader get the bearer Token from the header of the request
// success: returns the bearer token without the prefix
// error: return an empty token and an error
func GetBearerTokenFromHeader(r *http.Request, headerName string) (string, error) {
	qualifiedToken := r.Header.Get(headerName)
	hasPrefix := strings.HasPrefix(qualifiedToken, headerPrefix)
	if !hasPrefix {
		return "", errors.New(fmt.Sprintf("requested Header %q has not prefix %q : %q", headerName, headerPrefix, qualifiedToken))
	}
	return strings.TrimPrefix(qualifiedToken, headerPrefix), nil
}

// ContextWithTokenKey creates a new Context with the token
func ContextWithTokenKey(targetCtx context.Context, token Token) context.Context {
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

// GetAuthHeader Creates the valid value for the authentication header
// based on a context that already contains a  authorization token
func GetAuthHeader(ctx context.Context) (string, bool) {
	tok, ok := GetTokenFromContext(ctx)
	if !ok {
		return "", false
	}
	return headerPrefix + tok.GetValue(), true
}
