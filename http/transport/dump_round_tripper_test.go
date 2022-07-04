// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/05/20 by Vincent Landgraf

package transport

import (
	"bytes"
	"context"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/pace/bricks/maintenance/log"
	"github.com/pace/bricks/pkg/redact"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDumpRoundTripperEnv(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := log.Output(out).WithContext(context.Background())

	require.NotPanics(t, func() {
		rt := NewDumpRoundTripperEnv()
		assert.NotNil(t, rt)

		req := httptest.NewRequest("GET", "/foo", nil)
		req = req.WithContext(ctx)
		rt.SetTransport(&transportWithResponse{})

		_, err := rt.RoundTrip(req)
		assert.NoError(t, err)

		assert.Equal(t, "", out.String())
	})
}

func TestNewDumpRoundTripperEnvDisablePrefixBasedComplete(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := log.Output(out).WithContext(context.Background())

	require.NotPanics(t, func() {
		defer os.Setenv("HTTP_TRANSPORT_DUMP_DISABLE_ALL_URL_PREFIX", os.Getenv("HTTP_TRANSPORT_DUMP_DISABLE_ALL_URL_PREFIX"))
		os.Setenv("HTTP_TRANSPORT_DUMP_DISABLE_ALL_URL_PREFIX", "https://please-ignore-me")
		rt, err := NewDumpRoundTripper(
			roundTripConfigViaEnv(),
			RoundTripConfig(
				DumpRoundTripperOptionRequest,
				DumpRoundTripperOptionRequestHEX,
				DumpRoundTripperOptionResponse,
				DumpRoundTripperOptionResponseHEX,
				DumpRoundTripperOptionBody,
			))
		require.NoError(t, err)
		assert.NotNil(t, rt)

		req := httptest.NewRequest("GET", "/foo", bytes.NewBufferString("Foo"))
		req = req.WithContext(ctx)
		rt.SetTransport(&transportWithResponse{})

		_, err = rt.RoundTrip(req)
		assert.NoError(t, err)

		assert.Contains(t, out.String(), `"level":"debug"`)
		assert.Contains(t, out.String(), `"request":"GET /foo HTTP/1.1\r\nHost: example.com\r\n\r\nFoo"`)
		assert.Contains(t, out.String(), `"request-hex":"474554202f666f6f20485454502f312e310d0a486f73743a206578616d706c652e636f6d0d0a0d0a466f6f"`)
		assert.Contains(t, out.String(), `"response":"HTTP/0.0 000 status code 0\r\nContent-Length: 0\r\n\r\n"`)
		assert.Contains(t, out.String(), `"response-hex":"485454502f302e30203030302073746174757320636f646520300d0a436f6e74656e742d4c656e6774683a20300d0a0d0a"`)
		assert.Contains(t, out.String(), `"message":"HTTP Transport Dump"`)

		out.Reset()

		assert.Equal(t, "", out.String())

		reqWithPrefix := httptest.NewRequest("GET", "https://please-ignore-me.org/foo/", bytes.NewBufferString("Foo"))
		reqWithPrefix = reqWithPrefix.WithContext(ctx)

		_, err = rt.RoundTrip(reqWithPrefix)
		assert.NoError(t, err)
		assert.Empty(t, out.String())
	})
}

func TestNewDumpRoundTripperEnvDisablePrefixBasedBody(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := log.Output(out).WithContext(context.Background())

	require.NotPanics(t, func() {
		defer os.Setenv("HTTP_TRANSPORT_DUMP_DISABLE_DUMP_BODY_URL_PREFIX", os.Getenv("HTTP_TRANSPORT_DUMP_DISABLE_DUMP_BODY_URL_PREFIX"))
		os.Setenv("HTTP_TRANSPORT_DUMP_DISABLE_DUMP_BODY_URL_PREFIX", "https://please-ignore-me")
		rt, err := NewDumpRoundTripper(
			roundTripConfigViaEnv(),
			RoundTripConfig(
				DumpRoundTripperOptionRequest,
				DumpRoundTripperOptionRequestHEX,
				DumpRoundTripperOptionResponse,
				DumpRoundTripperOptionResponseHEX,
				DumpRoundTripperOptionBody,
			))
		require.NoError(t, err)
		assert.NotNil(t, rt)

		req := httptest.NewRequest("GET", "/foo", bytes.NewBufferString("Foo"))
		req = req.WithContext(ctx)
		rt.SetTransport(&transportWithResponse{})

		_, err = rt.RoundTrip(req)
		assert.NoError(t, err)

		assert.Contains(t, out.String(), `"level":"debug"`)
		assert.Contains(t, out.String(), `"request":"GET /foo HTTP/1.1\r\nHost: example.com\r\n\r\nFoo"`)
		assert.Contains(t, out.String(), `"request-hex":"474554202f666f6f20485454502f312e310d0a486f73743a206578616d706c652e636f6d0d0a0d0a466f6f"`)
		assert.Contains(t, out.String(), `"response":"HTTP/0.0 000 status code 0\r\nContent-Length: 0\r\n\r\n"`)
		assert.Contains(t, out.String(), `"response-hex":"485454502f302e30203030302073746174757320636f646520300d0a436f6e74656e742d4c656e6774683a20300d0a0d0a"`)
		assert.Contains(t, out.String(), `"message":"HTTP Transport Dump"`)

		out.Reset()

		assert.Equal(t, "", out.String())

		reqWithPrefix := httptest.NewRequest("GET", "https://please-ignore-me.org/foo/", bytes.NewBufferString("Foo"))
		reqWithPrefix = reqWithPrefix.WithContext(ctx)

		_, err = rt.RoundTrip(reqWithPrefix)
		assert.NoError(t, err)

		assert.Contains(t, out.String(), `"level":"debug"`)
		assert.Contains(t, out.String(), `"request":"GET https://please-ignore-me.org/foo/ HTTP/1.1\r\n\r\n"`)
		assert.Contains(t, out.String(), `"request-hex":"4745542068747470733a2f2f706c656173652d69676e6f72652d6d652e6f72672f666f6f2f20485454502f312e310d0a0d0a"`)
		assert.Contains(t, out.String(), `"response":"HTTP/0.0 000 status code 0\r\nContent-Length: 0\r\n\r\n"`)
		assert.Contains(t, out.String(), `"response-hex":"485454502f302e30203030302073746174757320636f646520300d0a436f6e74656e742d4c656e6774683a20300d0a0d0a"`)
		assert.Contains(t, out.String(), `"message":"HTTP Transport Dump"`)
	})
}

func TestNewDumpRoundTripper(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := log.Output(out).WithContext(context.Background())

	rt, err := NewDumpRoundTripper(
		RoundTripConfig(
			DumpRoundTripperOptionRequest,
			DumpRoundTripperOptionRequestHEX,
			DumpRoundTripperOptionResponse,
			DumpRoundTripperOptionResponseHEX,
			DumpRoundTripperOptionBody,
		),
	)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/foo", bytes.NewBufferString("Foo"))
	req = req.WithContext(ctx)
	rt.SetTransport(&transportWithResponse{})

	_, err = rt.RoundTrip(req)
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

	rt, err := NewDumpRoundTripper(
		RoundTripConfig(
			DumpRoundTripperOptionRequest,
			DumpRoundTripperOptionResponse,
			DumpRoundTripperOptionBody,
		),
	)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/foo", bytes.NewBufferString("Foo DE12345678909876543210 bar"))
	ctx = redact.Default.WithContext(ctx)
	req = req.WithContext(ctx)
	rt.SetTransport(&transportWithResponse{})

	_, err = rt.RoundTrip(req)
	assert.NoError(t, err)

	assert.Contains(t, out.String(), `"level":"debug"`)
	assert.Contains(t, out.String(), `"request":"GET /foo HTTP/1.1\r\nHost: example.com\r\n\r\nFoo ******************3210 bar"`)
	assert.Contains(t, out.String(), `"response":"HTTP/0.0 000 status code 0\r\nContent-Length: 0\r\n\r\n"`)
	assert.Contains(t, out.String(), `"message":"HTTP Transport Dump"`)
}

func TestNewDumpRoundTripperRedactedBasicAuth(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := log.Output(out).WithContext(context.Background())

	rt, err := NewDumpRoundTripper(
		RoundTripConfig(
			DumpRoundTripperOptionRequest,
			DumpRoundTripperOptionResponse,
			DumpRoundTripperOptionBody,
		),
	)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/foo", bytes.NewBufferString("Authorization: Basic ZGVtbzpwQDU1dzByZA=="))
	ctx = redact.Default.WithContext(ctx)
	req = req.WithContext(ctx)
	rt.SetTransport(&transportWithResponse{})

	_, err = rt.RoundTrip(req)
	assert.NoError(t, err)

	assert.Contains(t, out.String(), `"level":"debug"`)
	assert.Contains(t, out.String(), `"request":"GET /foo HTTP/1.1\r\nHost: example.com\r\n\r\n*************************************ZA=="`)
	assert.Contains(t, out.String(), `"response":"HTTP/0.0 000 status code 0\r\nContent-Length: 0\r\n\r\n"`)
	assert.Contains(t, out.String(), `"message":"HTTP Transport Dump"`)
}

func TestNewDumpRoundTripperSimple(t *testing.T) {
	out := &bytes.Buffer{}
	ctx := log.Output(out).WithContext(context.Background())

	rt, err := NewDumpRoundTripper(
		RoundTripConfig(
			DumpRoundTripperOptionRequest,
			DumpRoundTripperOptionResponse,
		),
	)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/foo", bytes.NewBufferString("Foo"))
	req = req.WithContext(ctx)
	rt.SetTransport(&transportWithResponse{})

	_, err = rt.RoundTrip(req)
	assert.NoError(t, err)

	assert.Contains(t, out.String(), `"level":"debug"`)
	assert.Contains(t, out.String(), `"request":"GET /foo HTTP/1.1\r\nHost: example.com\r\n\r\n"`)
	assert.Contains(t, out.String(), `"response":"HTTP/0.0 000 status code 0\r\nContent-Length: 0\r\n\r\n"`)
	assert.Contains(t, out.String(), `"message":"HTTP Transport Dump"`)
}
