// Package oauth2 provides a middelware that introspects the auth token on
// behalf of PACE services and populate the request context with useful information
// when the token is valid, otherwise aborts the request.

// TODO
// i would return 502 in case the upstream is not working
// introspec in new file
// table tests.
package oauth2

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type ctxkey string

var tokenKey = ctxkey("Token")

const headerPrefix = "Bearer "

type introspecter func(mdw *Middleware, token string, resp *introspectResponse) error

var errInvalidToken = errors.New("User token is invalid")
var errConnection = errors.New("problem connecting to the introspection endpoint")

// Oauth2 Middleware.
type Middleware struct {
	URL          string
	ClientID     string
	ClientSecret string
}

type introspectResponse struct {
	Status   bool    `json:"active"`
	Scope    *string `json:"scope"` // Could be string or nil, hence the pointer.
	ClientID string  `json:"client_id"`
	UserID   string  `json:"user_id"`
}

type token struct {
	value    string
	userID   string
	clientID string
	scopes   []string
}

// Should take token, introspect it, and put the token and other relevant information back
// in the context.
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		qualifiedToken := r.Header.Get("Authorization")

		items := strings.Split(qualifiedToken, "Bearer ")
		if len(items) < 2 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		receivedToken := items[1]
		var s introspectResponse
		err := introspect(*m, receivedToken, &s)
		log.Println(err)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		token := token{
			userID:   s.UserID,
			value:    receivedToken,
			clientID: s.ClientID,
		}

		if *s.Scope != "null" {
			scopes := strings.Split(*s.Scope, " ")
			token.scopes = scopes
		}

		ctx := context.WithValue(r.Context(), tokenKey, &token)
		next.ServeHTTP(w, r.WithContext(ctx))
		return
	})
}

func introspect(m Middleware, token string, s *introspectResponse) error {
	resp, err := http.PostForm(m.URL+"/oauth2/introspect",
		url.Values{"client_id": {m.ClientID}, "client_secret": {m.ClientSecret}, "token": {token}})

	if err != nil {
		log.Printf("%v\n", err)
		return errConnection
	}

	defer resp.Body.Close()

	// If Response is not 200, it means there are problems with setup, such
	// as wrong client ID or secret.
	if resp.StatusCode != 200 {
		log.Printf("Received %s from server, most likely bad oauth config.\n", resp.StatusCode)
		return errConnection
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(s)
	if err != nil {
		log.Printf("%v", err)
		return errConnection
	}

	if s.Status == false {
		return errInvalidToken
	}

	// Set the UserID of the introspect response manually since Cockpit returns
	// is in the response header and not the json (which we should change, I think).
	s.UserID = resp.Header.Get("X-UID")

	return nil
}

// TODO Pseudoish. To test.
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
