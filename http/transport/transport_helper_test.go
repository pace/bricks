// Copyright © 2020–2026 by PACE Mobility GmbH. All rights reserved.

package transport

import (
	"io"
	"net/http"
	"strings"
)

type transportWithResponse struct {
	statusCode int
	body       string
}

func (t *transportWithResponse) RoundTrip(req *http.Request) (*http.Response, error) {
	resp := &http.Response{StatusCode: t.statusCode}

	resp.Body = io.NopCloser(strings.NewReader(t.body))

	return resp, nil
}
