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

type Token interface {
	GetValue() string
}

type ctx string

// HeaderPrefix prefix of the Authentication value in the header
const HeaderPrefix = "Bearer "

var tokenKey = ctx("Token")

// GetBearerTokenFromHeader get the bearer Token from the header of the request
// success: returns the bearer token without the prefix
// error: return a empty token and a error
func GetBearerTokenFromHeader(r *http.Request, headerName string) (string, error) {
	qualifiedToken := r.Header.Get(headerName)
	hasPrefix := strings.HasPrefix(qualifiedToken, HeaderPrefix)
	if !hasPrefix {
		return "", errors.New(fmt.Sprintf("requested Header %q has not prefix %q : %q", headerName, HeaderPrefix, qualifiedToken))
	}
	return strings.TrimPrefix(qualifiedToken, HeaderPrefix), nil
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
