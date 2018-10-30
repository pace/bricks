// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/10/30 by Vincent Landgraf

// Client package contains a generic JSON API client for cockpit, cloud and jarvis
package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
	"lab.jamit.de/pace/go-microservice/maintenance/log"
)

// ErrNotFound error for HTTP 404 responses
var ErrNotFound = errors.New("Resource not found (HTTP 404)")

// Client implements a generic client for cockpit, cloud and jarvis
type Client struct {
	Endpoint     string // Endpoint URL
	Language     string // You can specify a locale (en|de)
	*http.Client        // http client used for doing the requests
}

// New creates a new Client with configured endpoint and the
// default http client
func New(endpoint string) *Client {
	return &Client{Endpoint: endpoint, Client: http.DefaultClient}
}

// URL generates a new URL for the given path and url values
func (c *Client) URL(path string, values url.Values) (*url.URL, error) {
	u, err := url.Parse(c.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("Endpoint URL %q can't be parsed: %v", c.Endpoint, err)
	}

	u.Path = path
	u.RawQuery = values.Encode()
	return u, nil
}

// GetJSON gets the json from the passed url and decodes its response into the given value v
func (c *Client) GetJSON(ctx context.Context, u *url.URL, v interface{}) error {
	uri := u.String()
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return err
	}

	return c.DoJSON(ctx, req, v)
}

// DoJSON executes a request and decodes its response it into the given value v
func (c *Client) DoJSON(ctx context.Context, req *http.Request, v interface{}) error {
	req = req.WithContext(ctx)
	req.Header.Set("Accept", "application/json")

	// set language if language given
	if c.Language != "" {
		req.Header.Set("Accept-Language", c.Language)
	}

	// add logging & tracing statement begin
	startTime := time.Now()
	span, _ := opentracing.StartSpanFromContext(ctx, "PACE DoJSON",
		opentracing.StartTime(startTime))
	span.LogFields(olog.String("url", req.URL.String()),
		olog.String("method", req.Method))

	// do actual request
	resp, err := c.Do(req)
	defer resp.Body.Close() // nolint: megacheck,errcheck

	// add logging & tracing statement end
	dur := float64(time.Since(startTime)) / float64(time.Millisecond)
	le := log.Ctx(ctx).Debug().
		Str("url", req.URL.String()).
		Str("method", req.Method).
		Float64("duration", dur)

	// add error or result set info
	if err != nil {
		le = le.Err(err)
		span.LogFields(olog.Error(err))
	}

	if resp != nil {
		le = le.Int("code", resp.StatusCode)
		span.LogFields(olog.Int("code", resp.StatusCode))
	}

	le.Msg("PACE DoJSON")
	span.Finish()

	// handle error if request failed
	if err != nil {
		return fmt.Errorf("Failed to %s %q: %v", req.Method, req.URL.String(), err)
	}

	// return not found
	if resp.StatusCode == 404 {
		return ErrNotFound
	}

	// decode response
	return json.NewDecoder(resp.Body).Decode(v)
}
