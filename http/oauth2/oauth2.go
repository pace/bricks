// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/03 by Mohamed Wael Khobalatte

// Package oauth2 provides a middelware that introspects the auth token on
// behalf of PACE services and populate the request context with useful information
// when the token is valid, otherwise aborts the request.
package oauth2

import (
	"context"
	"errors"
	"net/http"

	opentracing "github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"

	"github.com/pace/bricks/http/security"
	"github.com/pace/bricks/maintenance/log"
)

// Deprecated: Middleware holds data necessary for Oauth processing - Deprecated for generated apis,
// use the generated Authentication Backend of the API with oauth2.Authorizer
type Middleware struct {
	Backend TokenIntrospecter
}

// Deprecated: NewMiddleware creates a new Oauth middleware - Deprecated for generated apis,
// use the generated AuthenticationBackend of the API with oauth2.Authorizer
func NewMiddleware(backend TokenIntrospecter) *Middleware {
	return &Middleware{Backend: backend}
}

// Handler will parse the bearer token, introspect it, and put the token and other
// relevant information back in the context.
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, isOk := introspectRequest(r, w, m.Backend)
		if !isOk {
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type token struct {
	value    string
	userID   string
	clientID string
	scope    Scope
	backend  interface{}
}

const oAuth2Header = "Authorization"

// GetValue returns the oauth2 token of the current user
func (t *token) GetValue() string {
	return t.value
}

// IntrospectRequest introspects the requests and handles all errors:
// Success: it returns a context containing the introspection result and true
// if the introspection was successful
// Error: The function writes the error in the Response and creates a log-message
// with more details and returns nil and false if any error occurs during the introspection
func introspectRequest(r *http.Request, w http.ResponseWriter, tokenIntro TokenIntrospecter) (context.Context, bool) {
	// Setup tracing
	span, ctx := opentracing.StartSpanFromContext(r.Context(), "Oauth2")
	defer span.Finish()

	tok := security.GetBearerTokenFromHeader(r.Header.Get(oAuth2Header))
	if tok == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil, false
	}
	s, err := tokenIntro.IntrospectToken(ctx, tok)
	if err != nil {
		if errors.Is(err, ErrBadUpstreamResponse) || errors.Is(err, ErrUpstreamConnection) {
			http.Error(w, err.Error(), http.StatusBadGateway)
		} else if errors.Is(err, ErrInvalidToken) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		log.Req(r).Info().Msg(err.Error())
		return nil, false
	}
	t := fromIntrospectResponse(s, tok)
	ctx = security.ContextWithToken(ctx, &t)
	log.Req(r).Info().
		Str("client_id", t.clientID).
		Str("user_id", t.userID).
		Msg("Oauth2")
	span.LogFields(olog.String("client_id", t.clientID), olog.String("user_id", t.userID))
	return ctx, true
}

func fromIntrospectResponse(s *IntrospectResponse, tokenValue string) token {
	t := token{
		userID:   s.UserID,
		value:    tokenValue,
		clientID: s.ClientID,
		backend:  s.Backend,
	}

	t.scope = Scope(s.Scope)
	return t
}

// Request adds Authorization token to r
func Request(r *http.Request) *http.Request {
	tok, ok := security.GetTokenFromContext(r.Context())
	if ok {
		r.Header.Set("Authorization", security.GetAuthHeader(tok))
	}

	return r
}

// HasScope extracts an access token from context and checks if
// the permissions represented by the provided scope are included in the valid scope.
func HasScope(ctx context.Context, scope Scope) bool {
	tok, _ := security.GetTokenFromContext(ctx)
	oauth2token, ok := tok.(*token)
	if !ok {
		return false
	}
	return scope.IsIncludedIn(oauth2token.scope)
}

// UserID returns the userID stored in ctx
func UserID(ctx context.Context) (string, bool) {
	tok, _ := security.GetTokenFromContext(ctx)
	oauth2token, ok := tok.(*token)
	if !ok {
		return "", false
	}
	return oauth2token.userID, true
}

// Scopes returns the scopes stored in ctx
func Scopes(ctx context.Context) []string {
	tok, _ := security.GetTokenFromContext(ctx)
	oauth2token, ok := tok.(*token)
	if !ok {
		return []string{}
	}
	return oauth2token.scope.toSlice()
}

// ClientID returns the clientID stored in ctx
func ClientID(ctx context.Context) (string, bool) {
	tok, _ := security.GetTokenFromContext(ctx)
	oauth2token, ok := tok.(*token)
	if !ok {
		return "", false
	}
	return oauth2token.clientID, true

}

// Backend returns the backend stored in the context. It identifies the
// authorization backend for the token.
func Backend(ctx context.Context) (interface{}, bool) {
	tok, _ := security.GetTokenFromContext(ctx)
	oauth2token, ok := tok.(*token)
	if !ok {
		return nil, false
	}
	return oauth2token.backend, true
}

// ContextTransfer sources the oauth2 token from the sourceCtx
// and returning a new context based on the targetCtx
func ContextTransfer(sourceCtx context.Context, targetCtx context.Context) context.Context {
	tok, _ := security.GetTokenFromContext(sourceCtx)
	return security.ContextWithToken(targetCtx, tok)
}

// Deprecated: BearerToken was moved to the security package,
// because it's used by apiKey and oauth2 authorization.
// BearerToken returns the bearer token stored in ctx
func BearerToken(ctx context.Context) (string, bool) {
	if tok, ok := security.GetTokenFromContext(ctx); ok {
		return tok.GetValue(), true
	}
	return "", false
}

// Deprecated: WithBearerToken was moved to the security package,
// because it's used by api key and oauth2 authorization.
// returns a new context with the given bearer token
// Use security.BearerToken() to retrieve the token. Use Request() to obtain a request
// with the Authorization header set accordingly.
func WithBearerToken(ctx context.Context, bearerToken string) context.Context {
	return security.ContextWithToken(ctx, &token{value: bearerToken})
}
