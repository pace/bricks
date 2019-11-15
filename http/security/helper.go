// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/30 by Charlotte Pröller

package security

import (
	"context"
	"net/http"
	"strings"
)

type Token interface {
	GetValue() string
}

type bearerToken struct {
	value string
}

type ctxkey string

const HeaderPrefix = "Bearer "

var tokenKey = ctxkey("Token")

// GetValue returns the be
func (b *bearerToken) GetValue() string {
	return b.value
}

// GetBearerTokenFromHeader get the bearer Token from the header of the request
// success: returns the bearer token without the Prefix
// error: write the error in the response and return a empty bearer
func GetBearerTokenFromHeader(r *http.Request, w http.ResponseWriter, headerName string) string {
	qualifiedToken := r.Header.Get(headerName)
	items := strings.Split(qualifiedToken, HeaderPrefix)
	if len(items) < 2 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return ""
	}
	tokenValue := items[1]
	return tokenValue
}

// ContextWithTokenKey creates a context with a Token
func ContextWithTokenKey(targetCtx context.Context, token Token) context.Context {
	return context.WithValue(targetCtx, tokenKey, token)
}

// WithBearerToken returns a new context that has the given bearer token set.
// Use BearerToken() to retrieve the token. Use Request() to obtain a request
// with the Authorization header set accordingly.
func WithBearerToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenKey, &bearerToken{value: token})
}

// BearerToken returns the bearer token stored in ctx
func BearerToken(ctx context.Context) (string, bool) {
	token := TokenFromContext(ctx)

	if token == nil {
		return "", false
	}

	return token.GetValue(), true
}

func TokenFromContext(ctx context.Context) Token {
	val := ctx.Value(tokenKey)

	if val == nil {
		return nil
	}

	return val.(Token)
}
