// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/10/30 by Vincent Landgraf

// Package client contains a generic JSON API client for cockpit, cloud and jarvis
package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
	"golang.org/x/net/http2"
	"lab.jamit.de/pace/go-microservice/maintenance/log"
)

// ErrNotFound error for HTTP 404 responses
var ErrNotFound = errors.New("Resource not found (HTTP 404)")

// ErrRequest contains error details
type ErrRequest struct {
	ErrorDetail struct {
		Type  string `json:"type"`
		Codes []struct {
			Param  string `json:"param"`
			CodeID string `json:"code_id"`
		} `json:"codes"`
	} `json:"error"`
}

// Error implements the error interface
func (e *ErrRequest) Error() string {
	codes := fmt.Sprintf("Error request %s: ", e.ErrorDetail.Type)
	for _, code := range e.ErrorDetail.Codes {
		codes += fmt.Sprintf("Param %q: %q;", code.Param, code.CodeID)
	}
	return codes
}

// Client implements a generic client for cockpit, cloud and jarvis
type Client struct {
	Endpoint     string // Endpoint URL
	Language     string // You can specify a locale (en|de)
	Retries      int    // using exponential backoff
	*http.Client        // http client used for doing the requests
}

// New creates a new Client with configured endpoint and the
// default http client
func New(endpoint string) *Client {
	return &Client{Endpoint: endpoint, Client: http.DefaultClient, Retries: 10}
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
func (c *Client) GetJSON(ctx context.Context, url fmt.Stringer, v interface{}) error {
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return err
	}

	return c.DoJSON(ctx, req, v)
}

// DoJSON executes a request and decodes its response it into the given value v
func (c *Client) DoJSON(ctx context.Context, req *http.Request, v interface{}) error {
	// set language if language given
	if c.Language != "" {
		req.Header.Set("Accept-Language", c.Language)
	}
	req.Header.Set("Accept", "application/json")

	tries := 0
req:
	resp, err := c.Do(ctx, req)

	// if the err it temporary or the server returned a 500, 502 the
	// request will be retried
	uerr, ok := err.(*url.Error)
	gaerr, _ := err.(*http2.GoAwayError)

	if (ok && (uerr.Temporary() || uerr.Timeout())) || gaerr != nil ||
		(resp != nil && (resp.StatusCode == 500 || resp.StatusCode == 502)) {
		if tries < c.Retries {
			tries++
			seconds := time.Second * time.Duration(math.Pow(2, float64(tries)/2))
			log.Ctx(ctx).Info().Int("retries", tries).Err(err).Msgf("Received temporary error, will retry request after %v", seconds)
			time.Sleep(seconds)
			goto req
		}

		log.Ctx(ctx).Warn().Err(err).Msg("Too many retires, giving up")
	}
	if err != nil {
		return err
	}

	defer resp.Body.Close() // nolint: megacheck,errcheck

	// decode response
	return json.NewDecoder(resp.Body).Decode(v)
}

// Do executes a request and returns its response while logging and tracing it
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	cmdName := fmt.Sprintf("HTTP %s %s", req.Method, req.URL.Host)

	// add logging & tracing statement begin
	startTime := time.Now()
	span, _ := opentracing.StartSpanFromContext(ctx, cmdName,
		opentracing.StartTime(startTime))
	span.LogFields(olog.String("url", req.URL.String()),
		olog.String("method", req.Method))

	// do actual request
	resp, err := c.Client.Do(req)

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

	le.Msg(cmdName)
	span.Finish()

	// handle error if request failed
	if err != nil {
		return nil, fmt.Errorf("Failed to %s %q: %v", req.Method, req.URL.String(), err)
	}

	// return not found
	if resp.StatusCode == 404 {
		return nil, ErrNotFound
	}

	// return other error
	if resp.StatusCode == 400 {
		var errReq ErrRequest
		err := json.NewDecoder(resp.Body).Decode(&errReq)
		if err != nil {
			return nil, fmt.Errorf("Error parsing error response: %v", err)
		}
		return nil, &errReq
	}

	return resp, nil
}
