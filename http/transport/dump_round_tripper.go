// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package transport

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/caarlos0/env/v11"

	"github.com/pace/bricks/maintenance/log"
	"github.com/pace/bricks/pkg/redact"
)

// DumpRoundTripper dumps requests and responses in one log event.
// This is not part of te request logger to be able to filter dumps more easily.
type DumpRoundTripper struct {
	transport http.RoundTripper

	options DumpOptions

	blacklistAnyDumpPrefixes  []string
	blacklistBodyDumpPrefixes []string
}

type DumpRoundTripperOption func(rt *DumpRoundTripper) (*DumpRoundTripper, error)

type dumpRoundTripperConfig struct {
	Options                   []string `env:"HTTP_TRANSPORT_DUMP" envSeparator:"," envDefault:""`
	BlacklistBodyDumpPrefixes []string `env:"HTTP_TRANSPORT_DUMP_DISABLE_DUMP_BODY_URL_PREFIX" envSeparator:"," envDefault:""`
	BlacklistAnyDumpPrefixes  []string `env:"HTTP_TRANSPORT_DUMP_DISABLE_ALL_URL_PREFIX" envSeparator:"," envDefault:""`
}

func roundTripConfigViaEnv() DumpRoundTripperOption {
	return func(rt *DumpRoundTripper) (*DumpRoundTripper, error) {
		var cfg dumpRoundTripperConfig

		if err := env.Parse(&cfg); err != nil {
			return rt, fmt.Errorf("failed to parse dump round tripper environment: %w", err)
		}

		for _, option := range cfg.Options {
			if !isDumpOptionValid(option) {
				return nil, fmt.Errorf("invalid dump option %q", option)
			}

			rt.options[option] = true
		}

		rt.blacklistAnyDumpPrefixes = cfg.BlacklistAnyDumpPrefixes
		rt.blacklistBodyDumpPrefixes = cfg.BlacklistBodyDumpPrefixes

		return rt, nil
	}
}

func RoundTripConfig(dumpOptions ...string) DumpRoundTripperOption {
	return func(rt *DumpRoundTripper) (*DumpRoundTripper, error) {
		for _, option := range dumpOptions {
			if !isDumpOptionValid(option) {
				return nil, fmt.Errorf("invalid dump option %q", option)
			}

			rt.options[option] = true
		}

		return rt, nil
	}
}

// NewDumpRoundTripperEnv creates a new RoundTripper based on the configuration
// that is passed via environment variables.
func NewDumpRoundTripperEnv() *DumpRoundTripper {
	rt, err := NewDumpRoundTripper(roundTripConfigViaEnv())
	if err != nil {
		log.Fatalf("failed to setup NewDumpRoundTripperEnv: %v", err)
	}

	return rt
}

// NewDumpRoundTripper return the roundtripper with configured options.
func NewDumpRoundTripper(options ...DumpRoundTripperOption) (*DumpRoundTripper, error) {
	rt := &DumpRoundTripper{options: DumpOptions{}}

	var err error

	for _, option := range options {
		rt, err = option(rt)
		if err != nil {
			return rt, err
		}
	}

	return rt, nil
}

// Transport returns the RoundTripper to make HTTP requests.
func (l *DumpRoundTripper) Transport() http.RoundTripper {
	return l.transport
}

// SetTransport sets the RoundTripper to make HTTP requests.
func (l *DumpRoundTripper) SetTransport(rt http.RoundTripper) {
	l.transport = rt
}

// AnyEnabled returns true if any logging is enabled.
func (l *DumpRoundTripper) AnyEnabled() bool {
	return l.options.AnyEnabled(DumpRoundTripperOptionRequest, DumpRoundTripperOptionRequestHEX, DumpRoundTripperOptionResponse, DumpRoundTripperOptionResponseHEX)
}

func (l *DumpRoundTripper) ContainsBlacklistedPrefix(url *url.URL, blacklist []string) bool {
	if len(blacklist) == 0 {
		return false
	}

	for _, prefix := range blacklist {
		// TODO (juf): Do benchmark and compare against using pre-constructed prefix-tree
		if strings.HasPrefix(url.String(), prefix) {
			return true
		}
	}

	return false
}

// RoundTrip executes a single HTTP transaction via Transport().
func (l *DumpRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var redactor *redact.PatternRedactor

	options := mergeDumpOptions(l.options, DumpRoundTripperOptionsFromCtx(req.Context()))

	// check if the content redaction is enabled
	if !options.IsEnabled(DumpRoundTripperOptionNoRedact) {
		redactor = redact.Ctx(req.Context())
	}

	// fast path if logging is disabled
	if !options.AnyEnabled(DumpRoundTripperOptionRequest, DumpRoundTripperOptionRequestHEX, DumpRoundTripperOptionResponse, DumpRoundTripperOptionResponseHEX) {
		return l.Transport().RoundTrip(req)
	}

	if l.ContainsBlacklistedPrefix(req.URL, l.blacklistAnyDumpPrefixes) {
		return l.Transport().RoundTrip(req)
	}

	dumpBody := options.IsEnabled(DumpRoundTripperOptionBody) && !l.ContainsBlacklistedPrefix(req.URL, l.blacklistBodyDumpPrefixes)

	dl := log.Ctx(req.Context()).Debug()

	// request logging
	if options.AnyEnabled(DumpRoundTripperOptionRequest, DumpRoundTripperOptionRequestHEX) {
		reqDump, err := httputil.DumpRequest(req, dumpBody)
		if err != nil {
			reqDump = []byte(err.Error())
		}

		// in case a redactor is present, redact the content before logging
		if redactor != nil {
			reqDump = []byte(redactor.Mask(string(reqDump)))
		}

		if options.IsEnabled(DumpRoundTripperOptionRequest) {
			dl = dl.Bytes(DumpRoundTripperOptionRequest, reqDump)
		}

		if options.IsEnabled(DumpRoundTripperOptionRequestHEX) {
			dl = dl.Str(DumpRoundTripperOptionRequestHEX, hex.EncodeToString(reqDump))
		}
	}

	resp, err := l.Transport().RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// response logging
	if options.AnyEnabled(DumpRoundTripperOptionResponse, DumpRoundTripperOptionResponseHEX) {
		respDump, err := httputil.DumpResponse(resp, dumpBody)
		if err != nil {
			respDump = []byte(err.Error())
		}

		// in case a redactor is present, redact the content before logging
		if redactor != nil {
			respDump = []byte(redactor.Mask(string(respDump)))
		}

		if options.IsEnabled(DumpRoundTripperOptionResponse) {
			dl = dl.Bytes(DumpRoundTripperOptionResponse, respDump)
		}

		if options.IsEnabled(DumpRoundTripperOptionResponseHEX) {
			dl = dl.Str(DumpRoundTripperOptionResponseHEX, hex.EncodeToString(respDump))
		}
	}

	// emit log
	dl.Msg("HTTP Transport Dump")

	return resp, err
}
