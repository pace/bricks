// Copyright © 2019–2026 by PACE Mobility GmbH. All rights reserved.

package transport

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestSourceRoundTripper(t *testing.T) {
	req := httptest.NewRequest("GET", "/foo", nil)

	rt := RequestSourceRoundTripper{SourceName: "foobar"}
	rt.SetTransport(&transportWithResponse{})

	_, err := rt.RoundTrip(req)
	assert.NoError(t, err)
	assert.Equal(t, []string{"foobar"}, req.Header["Request-Source"])
}
