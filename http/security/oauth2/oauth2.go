// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/03 by Mohamed Wael Khobalatte

// Package oauth2 provides a middelware that introspects the auth token on
// behalf of PACE services and populate the request context with useful information
// when the token is valid, otherwise aborts the request.
package oauth2

import (
	"context"
	"fmt"
	"net/http"

	"github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
	"github.com/pace/bricks/http/security"
	"github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/log"
)

type token struct {
	value    string
	userID   string
	clientID string
	scope    Scope
}

// Middleware holds data necessary for Oauth processing
type Middleware struct {
	Backend TokenIntrospector
}

func (t *token) GetValue() string {
	return t.value
}

// NewMiddleware creates a new Oauth middleware
func NewMiddleware(backend TokenIntrospector) *Middleware {
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

// Authenticate authenticates a request based on the given authenticator the config and the scope value
// Used when no middleware is used and the authentication is generated
// This methods does all required checking of the Authorization, to make the generated code as short as possible
// Input:
// authorizer: Should be a oauth2.Authorizer, is nil-checked and converted in this method.
// Success: returns the request with a context and true
// Error: returns the request without any changes and adds the error to the response
// false if the Authorizer was no OAuth2 Authorizer or any other error occures. The error is directly added
// to the response
func Authenticate(authorizer security.Authorizer, r *http.Request, w http.ResponseWriter, authConfig interface{}, scope string) (*http.Request, bool) {
	auth, ok := authorizer.(*Authorizer)
	if !ok {
		http.Error(w, errors.New("authentication configuration missing").Error(), http.StatusUnauthorized)
		return r, false
	}
	ctx, isOk := auth.WithScope(scope).Authorize(authConfig, r, w)

	// Check if authorisation was successful
	if !isOk {
		return r, false
	}
	return r.WithContext(ctx), true
}

// IntrospectRequest introspects the requests and handles all errors:
// Success: it returns a context containing the introspection result and true
// if the introspection was successful
// Error: The function writes the error in the Response and creates a log-message
// with more details and returns nil and false if any error occurs during the introspection
func introspectRequest(r *http.Request, w http.ResponseWriter, tokenIntro TokenIntrospector) (context.Context, bool) {
	// Setup tracing
	span, ctx := opentracing.StartSpanFromContext(r.Context(), "Oauth2")
	defer span.Finish()
	tokenValue := security.GetBearerTokenFromHeader(r, w, "Authorization")
	if tokenValue == "" {
		return nil, false
	}
	s, err := tokenIntro.IntrospectToken(ctx, tokenValue)
	if err != nil {
		switch err {
		case ErrBadUpstreamResponse:
			log.Req(r).Info().Msg(err.Error())
			http.Error(w, err.Error(), http.StatusBadGateway)
			return nil, false
		case ErrUpstreamConnection:
			log.Req(r).Info().Msg(err.Error())
			http.Error(w, err.Error(), http.StatusBadGateway)
			return nil, false
		case ErrInvalidToken:
			log.Req(r).Info().Msg(err.Error())
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return nil, false
		}
	}
	t := fromIntrospectResponse(s, tokenValue)
	ctx = context.WithValue(ctx, security.TokenKey, &t)
	log.Req(r).Info().
		Str("client_id", t.clientID).
		Str("user_id", t.userID).
		Msg("Oauth2")
	span.LogFields(olog.String("client_id", t.clientID), olog.String("user_id", t.userID))
	return ctx, false
}

// ValidateScope  Validates based on the context if the scope is matching the required scope for this request
// Success: return true
// Error: error is written in the Response and the function returns false
func validateScope(ctx context.Context, w http.ResponseWriter, requiredScope string) bool {
	if !HasScope(ctx, Scope(requiredScope)) {
		http.Error(w, fmt.Sprintf("Forbidden - requires scope %q", requiredScope), http.StatusForbidden)
		return false
	}
	return true
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
	bt, ok := security.BearerToken(r.Context())

	if !ok {
		return r
	}

	authHeaderVal := security.HeaderPrefix + bt
	r.Header.Set("Authorization", authHeaderVal)
	return r
}

// HasScope extracts an access token T from context and checks if
// the permissions represented by the provided scope are included in T.
func HasScope(ctx context.Context, scope Scope) bool {
	tok := security.TokenFromContext(ctx)

	if tok == nil {
		return false
	}
	oauth2token, ok := tok.(*token)
	if !ok {
		return false
	}
	return scope.IsIncludedIn(oauth2token.scope)
}

// UserID returns the userID stored in ctx
func UserID(ctx context.Context) (string, bool) {
	tok := security.TokenFromContext(ctx)

	if tok == nil {
		return "", false
	}
	oauth2token, ok := tok.(*token)
	if !ok {
		return "", false
	}
	return oauth2token.userID, true
}

// Scopes returns the scopes stored in ctx
func Scopes(ctx context.Context) []string {
	tok := security.TokenFromContext(ctx)

	if tok == nil {
		return []string{}
	}
	oauth2token, ok := tok.(*token)
	if !ok {
		return []string{}
	}
	return oauth2token.scope.toSlice()
}

// ClientID returns the clientID stored in ctx
func ClientID(ctx context.Context) (string, bool) {
	tok := security.TokenFromContext(ctx)

	if tok == nil {
		return "", false
	}
	oauth2token, ok := tok.(*token)
	if !ok {
		return "", false
	}
	return oauth2token.clientID, true

}

// ContextTransfer sources the oauth2 token from the sourceCtx
// and returning a new context based on the targetCtx
func ContextTransfer(sourceCtx context.Context, targetCtx context.Context) context.Context {
	token := security.TokenFromContext(sourceCtx)
	return context.WithValue(targetCtx, security.TokenKey, token)
}
