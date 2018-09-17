// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/03 by Mohamed Wael Khobalatte

// Package oauth2 provides a middelware that introspects the auth token on
// behalf of PACE services and populate the request context with useful information
// when the token is valid, otherwise aborts the request.
//
// See example_usage.go for an example usage (pardon the runny wording).
package oauth2

import (
	"context"
	"net/http"
	"strings"
)

type ctxkey string

var tokenKey = ctxkey("Token")

const headerPrefix = "Bearer "

// Oauth2 Middleware.
type Middleware struct {
	URL            string
	ClientID       string
	ClientSecret   string
	introspectFunc introspecter
}

type token struct {
	value    string
	userID   string
	clientID string
	scopes   []string
}

// Handler will parse the bearer token, introspect it, and put the token and other
// relevant information back in the context.
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		qualifiedToken := r.Header.Get("Authorization")

		items := strings.Split(qualifiedToken, headerPrefix)
		if len(items) < 2 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenValue := items[1]
		var s introspectResponse
		var introspectErr error

		if m.introspectFunc != nil {
			introspectErr = m.introspectFunc(m, tokenValue, &s)
		} else {
			introspectErr = introspect(m, tokenValue, &s)
		}

		switch introspectErr {
		case errBadUpstreamResponse:
			http.Error(w, introspectErr.Error(), http.StatusBadGateway)
		case errUpstreamConnection:
			http.Error(w, introspectErr.Error(), http.StatusUnauthorized)
		case errInvalidToken:
			http.Error(w, introspectErr.Error(), http.StatusUnauthorized)
		}

		token := fromIntrospectResponse(s, tokenValue)

		ctx := context.WithValue(r.Context(), tokenKey, &token)
		next.ServeHTTP(w, r.WithContext(ctx))
		return
	})
}

func (m *Middleware) addIntrospectFunc(f introspecter) {
	m.introspectFunc = f
}

func fromIntrospectResponse(s introspectResponse, tokenValue string) token {
	token := token{
		userID:   s.UserID,
		value:    tokenValue,
		clientID: s.ClientID,
	}

	if s.Scope != "" {
		scopes := strings.Split(s.Scope, " ")
		token.scopes = scopes
	}

	return token
}

func Request(ctx context.Context, r *http.Request) *http.Request {
	token := BearerToken(ctx)
	authHeaderVal := headerPrefix + token
	r.Header.Set("Authorization: ", authHeaderVal)
	return r
}

func BearerToken(ctx context.Context) string {
	token := ctx.Value(tokenKey).(*token)
	return token.value
}

func HasScope(ctx context.Context, scope string) bool {
	token := ctx.Value(tokenKey).(*token)

	for _, v := range token.scopes {
		if v == scope {
			return true
		}
	}

	return false
}

func UserID(ctx context.Context) string {
	token := ctx.Value(tokenKey).(*token)

	return token.userID
}

func Scopes(ctx context.Context) []string {
	token := ctx.Value(tokenKey).(*token)

	return token.scopes
}

func ClientID(ctx context.Context) string {
	token := ctx.Value(tokenKey).(*token)

	return token.clientID
}
