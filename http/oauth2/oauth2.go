// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/03 by Mohamed Wael Khobalatte

// Package oauth2 provides a middelware that introspects the auth token on
// behalf of PACE services and populate the request context with useful information
// when the token is valid, otherwise aborts the request.
package oauth2

import (
	"context"
	"github.com/caarlos0/env"
	opentracing "github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
	"github.com/pace/bricks/maintenance/log"
	"net/http"
	"strings"
)

type ctxkey string

var tokenKey = ctxkey("Token")

const headerPrefix = "Bearer "

// Middleware holds data necessary for Oauth processing
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

type config struct {
	URL          string `env:"OAUTH2_URL" envDefault:"https://cp-1-prod.pacelink.net"`
	ClientID     string `env:"OAUTH2_CLIENT_ID"`
	ClientSecret string `env:"OAUTH2_CLIENT_SECRET"`
}

// NewMiddleware creates a new Oauth middleware
func NewMiddleware() *Middleware {
	var cfg config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Failed to parse oauth2 environment: %v", err)
	}
	return &Middleware{
		URL:          cfg.URL,
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
	}
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
		var s introspectResponse
		var introspectErr error

		if m.introspectFunc != nil {
			introspectErr = m.introspectFunc(m, tokenValue, &s)
		} else {
			introspectErr = introspect(m, tokenValue, &s)
		}

		switch introspectErr {
		case errBadUpstreamResponse:
			log.Req(r).Info().Msg(introspectErr.Error())
			http.Error(w, introspectErr.Error(), http.StatusBadGateway)
			return
		case errUpstreamConnection:
			log.Req(r).Info().Msg(introspectErr.Error())
			http.Error(w, introspectErr.Error(), http.StatusUnauthorized)
			return
		case errInvalidToken:
			log.Req(r).Info().Msg(introspectErr.Error())
			http.Error(w, introspectErr.Error(), http.StatusUnauthorized)
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

func (m *Middleware) addIntrospectFunc(f introspecter) {
	m.introspectFunc = f
}

func fromIntrospectResponse(s introspectResponse, tokenValue string) token {
	t := token{
		userID:   s.UserID,
		value:    tokenValue,
		clientID: s.ClientID,
	}

	if s.Scope != "" {
		scopes := strings.Split(s.Scope, " ")
		t.scopes = scopes
	}

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

// HasScope checks if scope is stored in ctx
func HasScope(ctx context.Context, scope string) bool {
	token := tokenFromContext(ctx)

	if token == nil {
		return false
	}

	for _, v := range token.scopes {
		if v == scope {
			return true
		}
	}

	return false
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

	return token.scopes
}

// ClientID returns the clientID stored in ctx
func ClientID(ctx context.Context) (string, bool) {
	token := tokenFromContext(ctx)

	if token == nil {
		return "", false
	}

	return token.clientID, true

}

func tokenFromContext(ctx context.Context) *token {
	val := ctx.Value(tokenKey)

	if val == nil {
		return nil
	}

	return val.(*token)
}
