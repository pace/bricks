// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/04/29 by Marius Neugebauer

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
