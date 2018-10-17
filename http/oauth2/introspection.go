// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/12 by Mohamed Wael Khobalatte

package oauth2

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"lab.jamit.de/pace/go-microservice/maintenance/log"
)

type introspecter func(mdw *Middleware, token string, resp *introspectResponse) error

var errInvalidToken = errors.New("User token is invalid")
var errUpstreamConnection = errors.New("Problem connecting to the introspection endpoint")
var errBadUpstreamResponse = errors.New("Bad upstream response when introspecting token")

type introspectResponse struct {
	Active   bool   `json:"active"`
	Scope    string `json:"scope"`
	ClientID string `json:"client_id"`
	UserID   string `json:"user_id"`
}

func introspect(m *Middleware, token string, s *introspectResponse) error {
	resp, err := http.PostForm(m.URL+"/oauth2/introspect",
		url.Values{"client_id": {m.ClientID}, "client_secret": {m.ClientSecret}, "token": {token}})

	if err != nil {
		log.Printf("%v\n", err)
		return errUpstreamConnection
	}

	defer resp.Body.Close() // nolint: errcheck,gosec

	// If Response is not 200, it means there are problems with setup, such
	// as wrong client ID or secret.
	if resp.StatusCode != 200 {
		log.Printf("Received %d from server, most likely bad oauth config.\n", resp.StatusCode)
		return errBadUpstreamResponse
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(s)
	if err != nil {
		log.Printf("%v", err)
		return errBadUpstreamResponse
	}

	if !s.Active {
		return errInvalidToken
	}

	// Set the UserID of the introspect response manually since Cockpit returns
	// is in the response header and not the json (which we should change, I think).
	s.UserID = resp.Header.Get("X-UID")

	return nil
}
