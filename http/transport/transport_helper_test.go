// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/05/18 by Vincent Landgraf

package transport

import "net/http"

type transportWithResponse struct {
	statusCode int
}

func (t *transportWithResponse) RoundTrip(req *http.Request) (*http.Response, error) {
	resp := &http.Response{StatusCode: t.statusCode}

	return resp, nil
}
