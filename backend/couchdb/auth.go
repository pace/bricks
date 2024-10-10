// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package couchdb

import "net/http"

type AuthTransport struct {
	Username, Password string
	transport          http.RoundTripper
}

func (l *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(l.Username, l.Password)
	return l.transport.RoundTrip(req)
}

func (l *AuthTransport) Transport() http.RoundTripper {
	return l.transport
}

func (l *AuthTransport) SetTransport(rt http.RoundTripper) {
	l.transport = rt
}
