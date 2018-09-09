package oauth2

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Middleware struct {
	Host         string
	ClientID     string
	ClientSecret string
}

type IntrospectResponse struct {
	Status   bool    `json:"active"`
	Scope    *string `json:"scope"` // Could be string or nil, hence the pointer.
	ClientID string  `json:"client_id"`
}

// Should take token, introspect it, and put the token and other relevant information back
// in the context.
// TODO
//
// 1. Move error code to well named functions.
// 2. Refaactor code that puts things in the context.
func (m *Middleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		qualifiedToken := r.Header.Get("Authorization")

		items := strings.Split(qualifiedToken, "Bearer ")
		if len(items) < 2 {
			// TODO
			//
			// Should these responses actually follow JSON spec?
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token := items[1]

		resp, err := http.PostForm(m.Host+"/oauth2/introspect",
			url.Values{"client_id": {m.ClientID}, "client_secret": {m.ClientSecret}, "token": {token}})

		defer resp.Body.Close()

		if err != nil {
			// Possible upstream error, log and fail. Use log instead of fmt?
			fmt.Printf("%v\n", err)
			http.Error(w, "InternalServerError", http.StatusInternalServerError)
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("%v\n", err)
			http.Error(w, "InternalServerError", http.StatusInternalServerError)
			return
		}

		// If Response is not 200, it means there are problems with setup, such
		// as wrong client ID or secret.
		if resp.StatusCode != 200 {
			fmt.Printf("Received %s from server, most likely bad oauth config.\n", resp.StatusCode)

			http.Error(w, "InternalServerError", http.StatusInternalServerError)
			return
		}

		var s = new(IntrospectResponse)
		err = json.Unmarshal(body, &s)
		if err != nil {
			fmt.Printf("%v", err)
			http.Error(w, "InternalServerError", http.StatusInternalServerError)
			return
		}

		if s.Status == false {
			fmt.Printf("Token %s not active", token)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if s.Status == true {
			xuid := resp.Header.Get("X-UID")

			// Unlikely to happen, but we check anyway.
			if xuid == "" {
				http.Error(w, "InternalServerError", http.StatusInternalServerError)
				return
			}

			// TODO:
			//
			// We might need to consider possible collusions if we pick a string
			// like this.
			ctx := context.WithValue(r.Context(), "X-UID", xuid)
			ctx = context.WithValue(ctx, "authToken", token)

			// If token is associated with one or more scopes, put them in
			// the context as a slice.
			if *s.Scope != "null" {
				scopes := strings.Split(*s.Scope, " ")
				ctx = context.WithValue(ctx, "scopes", scopes)
			}

			// Add ClientID to the context.
			ctx = context.WithValue(ctx, "ClientID", s.ClientID)

			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
	})
}

// TODO Pseudoish. To test.
func Request(ctx context.Context, r *http.Request) *http.Request {
	token := BearerToken(ctx)
	authHeaderVal := "Bearer " + token
	r.Header.Set("Authorization: ", authHeaderVal)
	return r
}

// TODO: To be tested separately.
func BearerToken(ctx context.Context) string {
	return Get(ctx, "authToken")
}

// TODO: To be tested separately.
func HasScope(ctx context.Context, scope string) bool {
	scopes, ok := ctx.Value("scopes").([]string)

	if !ok {
		return false
	}

	for _, v := range scopes {
		if v == scope {
			return true
		}
	}

	return false
}

// TODO: To be tested separately.
//
// Should we propagate an error back to caller? Otherwise this might panic.
func Get(ctx context.Context, key string) string {
	return (ctx.Value(key)).(string)
}
