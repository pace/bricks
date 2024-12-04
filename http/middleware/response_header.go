// Copyright © 2021 by PACE Telematics GmbH. All rights reserved.

package middleware

import (
	"errors"
	"net/http"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
)

// ClientIDHeaderName name of the HTTP header that is used for reporting.
const (
	ClientIDHeaderName = "Client-ID"
)

var ErrEmptyAuthorizedParty = errors.New("authorized party is empty")

func ClientID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := r.Header.Get("Authorization")
		if strings.HasPrefix(value, "Bearer ") {
			var claim clientIDClaim

			_, _, err := new(jwt.Parser).ParseUnverified(value[7:], &claim)
			if err == nil {
				w.Header().Add(ClientIDHeaderName, claim.AuthorizedParty)
			}
		}

		next.ServeHTTP(w, r)
	})
}

type clientIDClaim struct {
	jwt.MapClaims
	AuthorizedParty string `json:"azp"`
}

func (c clientIDClaim) Valid() error {
	if c.AuthorizedParty == "" {
		return ErrEmptyAuthorizedParty
	}

	return nil
}
