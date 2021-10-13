// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/05/20 by Vincent Landgraf

package transport

import (
	"encoding/hex"
	"net/http"
	"net/http/httputil"

	"github.com/caarlos0/env"
	"github.com/pace/bricks/maintenance/log"
	"github.com/pace/bricks/pkg/redact"
)

// DumpRoundTripper dumps requests and responses in one log event.
// This is not part of te request logger to be able to filter dumps more easily
type DumpRoundTripper struct {
	transport http.RoundTripper

	DumpRequest     bool
	DumpResponse    bool
	DumpRequestHEX  bool
	DumpResponseHEX bool
	DumpBody        bool
	DumpNoRedact    bool
}

type DumpRoundTripperOption string

const (
	DumpRoundTripperOptionRequest     = "request"
	DumpRoundTripperOptionResponse    = "response"
	DumpRoundTripperOptionRequestHEX  = "request-hex"
	DumpRoundTripperOptionResponseHEX = "response-hex"
	DumpRoundTripperOptionBody        = "body"
	DumpRoundTripperOptionNoRedact    = "no-redact"
)

// NewDumpRoundTripperEnv creates a new RoundTripper based on the configuration
// that is passed via environment variables
func NewDumpRoundTripperEnv() *DumpRoundTripper {
	// parse config
	var cfg struct {
		Options []string `env:"HTTP_TRANSPORT_DUMP" envSeparator:"," envDefault:""`
	}
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Failed to parse dump round tripper environment: %v", err)
	}

	// create and configure rt
	return NewDumpRoundTripper(cfg.Options...)
}

// NewDumpRoundTripper return the roundtripper with configured options
func NewDumpRoundTripper(options ...string) *DumpRoundTripper {
	rt := &DumpRoundTripper{}
	for _, option := range options {
		switch option {
		case DumpRoundTripperOptionRequest:
			rt.DumpRequest = true
		case DumpRoundTripperOptionResponse:
			rt.DumpResponse = true
		case DumpRoundTripperOptionRequestHEX:
			rt.DumpRequestHEX = true
		case DumpRoundTripperOptionResponseHEX:
			rt.DumpResponseHEX = true
		case DumpRoundTripperOptionBody:
			rt.DumpBody = true
		case DumpRoundTripperOptionNoRedact:
			rt.DumpNoRedact = true
		default:
			log.Fatalf("Failed to parse dump round tripper options from env: %v", option)
		}
	}
	return rt
}

// Transport returns the RoundTripper to make HTTP requests
func (l *DumpRoundTripper) Transport() http.RoundTripper {
	return l.transport
}

// SetTransport sets the RoundTripper to make HTTP requests
func (l *DumpRoundTripper) SetTransport(rt http.RoundTripper) {
	l.transport = rt
}

// AnyEnabled returns true if any logging is enabled
func (l DumpRoundTripper) AnyEnabled() bool {
	return l.DumpRequest || l.DumpResponse || l.DumpRequestHEX || l.DumpResponseHEX
}

// RoundTrip executes a single HTTP transaction via Transport()
func (l *DumpRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var redactor *redact.PatternRedactor

	// check if the content redaction is enabled
	if !l.DumpNoRedact {
		redactor = redact.Ctx(req.Context())
	}

	// fast path if logging is disabled
	if !l.AnyEnabled() {
		return l.Transport().RoundTrip(req)
	}

	dl := log.Ctx(req.Context()).Debug()

	// request logging
	if l.DumpRequest || l.DumpRequestHEX {
		reqDump, err := httputil.DumpRequest(req, l.DumpBody)
		if err != nil {
			reqDump = []byte(err.Error())
		}

		// in case a redactor is present, redact the content before logging
		if redactor != nil {
			reqDump = []byte(redactor.Mask(string(reqDump)))
		}

		if l.DumpRequest {
			dl = dl.Bytes(DumpRoundTripperOptionRequest, reqDump)
		}
		if l.DumpRequestHEX {
			dl = dl.Str(DumpRoundTripperOptionRequestHEX, hex.EncodeToString(reqDump))
		}
	}

	resp, err := l.Transport().RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// response logging
	if l.DumpResponse || l.DumpResponseHEX {
		respDump, err := httputil.DumpResponse(resp, l.DumpBody)
		if err != nil {
			respDump = []byte(err.Error())
		}

		// in case a redactor is present, redact the content before logging
		if redactor != nil {
			respDump = []byte(redactor.Mask(string(respDump)))
		}

		if l.DumpResponse {
			dl = dl.Bytes(DumpRoundTripperOptionResponse, respDump)
		}
		if l.DumpResponseHEX {
			dl = dl.Str(DumpRoundTripperOptionResponseHEX, hex.EncodeToString(respDump))
		}
	}

	// emit log
	dl.Msg("HTTP Transport Dump")

	return resp, err
}
