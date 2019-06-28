// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/03 by Mohamed Wael Khobalatte

// Package oauth2 provides a middelware that introspects the auth token on
// behalf of PACE services and populate the request context with useful information
// when the token is valid, otherwise aborts the request.
package oauth2

import (
	"context"
	"net/http"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
	"github.com/pace/bricks/maintenance/log"
)

type ctxkey string

var tokenKey = ctxkey("Token")

const headerPrefix = "Bearer "

// Middleware holds data necessary for Oauth processing
type Middleware struct {
	Backend TokenIntrospecter
}

type token struct {
	value    string
	userID   string
	clientID string
	scope    Scope
}

// NewMiddleware creates a new Oauth middleware
func NewMiddleware(backend TokenIntrospecter) *Middleware {
	return &Middleware{Backend: backend}
}

// Handler will parse the bearer token, introspect it, and put the token and other
// relevant information back in the context.
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Setup tracing
		span, ctx := opentracing.StartSpanFromContext(r.Context(), "Oauth2")
		defer span.Finish()

		qualifiedToken := r.Header.Get("Authorization")

		items := strings.Split(qualifiedToken, headerPrefix)
		if len(items) < 2 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenValue := items[1]

		s, err := m.Backend.IntrospectToken(ctx, tokenValue)
		switch err {
		case ErrBadUpstreamResponse:
			log.Req(r).Info().Msg(err.Error())
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		case ErrUpstreamConnection:
			log.Req(r).Info().Msg(err.Error())
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		case ErrInvalidToken:
			log.Req(r).Info().Msg(err.Error())
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		t := fromIntrospectResponse(s, tokenValue)

		ctx = context.WithValue(ctx, tokenKey, &t)

		log.Req(r).Info().
			Str("client_id", t.clientID).
			Str("user_id", t.userID).
			Msg("Oauth2")

		span.LogFields(olog.String("client_id", t.clientID), olog.String("user_id", t.userID))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func fromIntrospectResponse(s *IntrospectResponse, tokenValue string) token {
	t := token{
		userID:   s.UserID,
		value:    tokenValue,
		clientID: s.ClientID,
	}

	t.scope = Scope(s.Scope)
	return t
}

// Request adds Authorization token to r
func Request(r *http.Request) *http.Request {
	bt, ok := BearerToken(r.Context())

	if !ok {
		return r
	}

	authHeaderVal := headerPrefix + bt
	r.Header.Set("Authorization", authHeaderVal)
	return r
}

// BearerToken returns the bearer token stored in ctx
func BearerToken(ctx context.Context) (string, bool) {
	token := tokenFromContext(ctx)

	if token == nil {
		return "", false
	}

	return token.value, true
}

// HasScope extracts an access token T from context and checks if
// the permissions represented by the provided scope are included in T.
func HasScope(ctx context.Context, scope Scope) bool {
	token := tokenFromContext(ctx)

	if token == nil {
		return false
	}

	return scope.IsIncludedIn(token.scope)
}

// UserID returns the userID stored in ctx
func UserID(ctx context.Context) (string, bool) {
	token := tokenFromContext(ctx)

	if token == nil {
		return "", false
	}

	return token.userID, true
}

// Scopes returns the scopes stored in ctx
func Scopes(ctx context.Context) []string {
	token := tokenFromContext(ctx)

	if token == nil {
		return []string{}
	}

	return token.scope.toSlice()
}

// ClientID returns the clientID stored in ctx
func ClientID(ctx context.Context) (string, bool) {
	token := tokenFromContext(ctx)

	if token == nil {
		return "", false
	}

	return token.clientID, true

}

// ContextTransfer sources the oauth2 token from the sourceCtx
// and returning a new context based on the targetCtx
func ContextTransfer(sourceCtx context.Context, targetCtx context.Context) context.Context {
	token := tokenFromContext(sourceCtx)
	return context.WithValue(targetCtx, tokenKey, token)
}

func tokenFromContext(ctx context.Context) *token {
	val := ctx.Value(tokenKey)

	if val == nil {
		return nil
	}

	return val.(*token)
}
