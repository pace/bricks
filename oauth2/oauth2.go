// Package oauth2 provides a middelware that introspects the auth token on
// behalf of PACE services and populate the request context with useful information
// when the token is valid, otherwise aborts the request.
package oauth2

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
)

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
}

type User struct {
	authToken string
	userID    string
	clientID  string
	scopes    []string
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

		token := items[1]

		resp, err := m.introspect(token)
		defer resp.Body.Close()

		if err != nil {
			// Possible upstream error, log and fail. Use log instead of fmt?
			log.Printf("%v\n", err)
			http.Error(w, "InternalServerError", http.StatusInternalServerError)
			return
		}

		// If Response is not 200, it means there are problems with setup, such
		// as wrong client ID or secret.
		if resp.StatusCode != 200 {
			log.Printf("Received %s from server, most likely bad oauth config.\n", resp.StatusCode)

			http.Error(w, "InternalServerError", http.StatusInternalServerError)
			return
		}

		var s introspectResponse
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&s)
		if err != nil {
			log.Printf("%v", err)
			http.Error(w, "InternalServerError", http.StatusInternalServerError)
			return
		}

		if s.Status == false {
			log.Printf("Token %s not active", token)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if s.Status == true {
			xuid := resp.Header.Get("X-UID")

			// Unlikely to happen, but we check anyway.
			if xuid == "" {
				http.Error(w, "BadGateway", http.StatusBadGateway)
				return
			}

			user := User{
				userID:    xuid,
				authToken: token,
				clientID:  s.ClientID,
			}

			if *s.Scope != "null" {
				scopes := strings.Split(*s.Scope, " ")
				user.scopes = scopes
			}

			ctx := context.WithValue(r.Context(), "User", user)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
	})
}

func (m *Middleware) introspect(token string) (*http.Response, error) {
	return http.PostForm(m.URL+"/oauth2/introspect",
		url.Values{"client_id": {m.ClientID}, "client_secret": {m.ClientSecret}, "token": {token}})
}

// TODO Pseudoish. To test.
func Request(ctx context.Context, r *http.Request) *http.Request {
	token := BearerToken(ctx)
	authHeaderVal := "Bearer " + token
	r.Header.Set("Authorization: ", authHeaderVal)
	return r
}

func BearerToken(ctx context.Context) string {
	user := ctx.Value("User").(User)
	return user.authToken
}

func HasScope(ctx context.Context, scope string) bool {
	user := ctx.Value("User").(User)

	for _, v := range user.scopes {
		if v == scope {
			return true
		}
	}

	return false
}

func UserID(ctx context.Context) string {
	user := ctx.Value("User").(User)

	return user.userID
}

func Scopes(ctx context.Context) []string {
	user := ctx.Value("User").(User)

	return user.scopes
}

func ClientID(ctx context.Context) string {
	user := ctx.Value("User").(User)

	return user.clientID
}
