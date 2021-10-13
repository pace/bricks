// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/05/20 by Vincent Landgraf

package transport

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"

	"github.com/pace/bricks/maintenance/log"
	"github.com/pace/bricks/pkg/redact"
	"github.com/stretchr/testify/assert"
)

func TestNewDumpRoundTripperEnv(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := log.Output(out).WithContext(context.Background())

	rt := NewDumpRoundTripperEnv()
	assert.NotNil(t, rt)

	req := httptest.NewRequest("GET", "/foo", nil)
	req = req.WithContext(ctx)
	rt.SetTransport(&transportWithResponse{})

	_, err := rt.RoundTrip(req)
	assert.NoError(t, err)

	assert.Equal(t, "", out.String())
}

func TestNewDumpRoundTripper(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := log.Output(out).WithContext(context.Background())

	rt := NewDumpRoundTripper(
		DumpRoundTripperOptionRequest,
		DumpRoundTripperOptionRequestHEX,
		DumpRoundTripperOptionResponse,
		DumpRoundTripperOptionResponseHEX,
		DumpRoundTripperOptionBody,
	)

	req := httptest.NewRequest("GET", "/foo", bytes.NewBufferString("Foo"))
	req = req.WithContext(ctx)
	rt.SetTransport(&transportWithResponse{})

	_, err := rt.RoundTrip(req)
	assert.NoError(t, err)

	assert.Contains(t, out.String(), `"level":"debug"`)
	assert.Contains(t, out.String(), `"request":"GET /foo HTTP/1.1\r\nHost: example.com\r\n\r\nFoo"`)
	assert.Contains(t, out.String(), `"request-hex":"474554202f666f6f20485454502f312e310d0a486f73743a206578616d706c652e636f6d0d0a0d0a466f6f"`)
	assert.Contains(t, out.String(), `"response":"HTTP/0.0 000 status code 0\r\nContent-Length: 0\r\n\r\n"`)
	assert.Contains(t, out.String(), `"response-hex":"485454502f302e30203030302073746174757320636f646520300d0a436f6e74656e742d4c656e6774683a20300d0a0d0a"`)
	assert.Contains(t, out.String(), `"message":"HTTP Transport Dump"`)
}

func TestNewDumpRoundTripperRedacted(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := log.Output(out).WithContext(context.Background())

	rt := NewDumpRoundTripper(
		DumpRoundTripperOptionRequest,
		DumpRoundTripperOptionResponse,
		DumpRoundTripperOptionBody,
	)

	req := httptest.NewRequest("GET", "/foo", bytes.NewBufferString("Foo DE12345678909876543210 bar"))
	ctx = redact.Default.WithContext(ctx)
	req = req.WithContext(ctx)
	rt.SetTransport(&transportWithResponse{})

	_, err := rt.RoundTrip(req)
	assert.NoError(t, err)

	assert.Contains(t, out.String(), `"level":"debug"`)
	assert.Contains(t, out.String(), `"request":"GET /foo HTTP/1.1\r\nHost: example.com\r\n\r\nFoo ******************3210 bar"`)
	assert.Contains(t, out.String(), `"response":"HTTP/0.0 000 status code 0\r\nContent-Length: 0\r\n\r\n"`)
	assert.Contains(t, out.String(), `"message":"HTTP Transport Dump"`)
}

func TestNewDumpRoundTripperRedactedBasicAuth(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := log.Output(out).WithContext(context.Background())

	rt := NewDumpRoundTripper(
		DumpRoundTripperOptionRequest,
		DumpRoundTripperOptionResponse,
		DumpRoundTripperOptionBody,
	)

	req := httptest.NewRequest("GET", "/foo", bytes.NewBufferString("Authorization: Basic ZGVtbzpwQDU1dzByZA=="))
	ctx = redact.Default.WithContext(ctx)
	req = req.WithContext(ctx)
	rt.SetTransport(&transportWithResponse{})

	_, err := rt.RoundTrip(req)
	assert.NoError(t, err)

	assert.Contains(t, out.String(), `"level":"debug"`)
	assert.Contains(t, out.String(), `"request":"GET /foo HTTP/1.1\r\nHost: example.com\r\n\r\n*************************************ZA=="`)
	assert.Contains(t, out.String(), `"response":"HTTP/0.0 000 status code 0\r\nContent-Length: 0\r\n\r\n"`)
	assert.Contains(t, out.String(), `"message":"HTTP Transport Dump"`)
}

func TestNewDumpRoundTripperSimple(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := log.Output(out).WithContext(context.Background())

	rt := NewDumpRoundTripper(
		DumpRoundTripperOptionRequest,
		DumpRoundTripperOptionResponse,
	)

	req := httptest.NewRequest("GET", "/foo", bytes.NewBufferString("Foo"))
	req = req.WithContext(ctx)
	rt.SetTransport(&transportWithResponse{})

	_, err := rt.RoundTrip(req)
	assert.NoError(t, err)

	assert.Contains(t, out.String(), `"level":"debug"`)
	assert.Contains(t, out.String(), `"request":"GET /foo HTTP/1.1\r\nHost: example.com\r\n\r\n"`)
	assert.Contains(t, out.String(), `"response":"HTTP/0.0 000 status code 0\r\nContent-Length: 0\r\n\r\n"`)
	assert.Contains(t, out.String(), `"message":"HTTP Transport Dump"`)
}
