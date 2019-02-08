// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/10/12 by Vincent Landgraf

package nominatim

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
	"lab.jamit.de/pace/go-microservice/maintenance/log"
	"lab.jamit.de/pace/go-microservice/pkg/synctx"
)

// Client for nominatim using the configured endpoint
type Client struct {
	// Endpoint url e.g. https://nominatim.example.org/
	Endpoint string
}

// ErrUnableToGeocode lat/lon don't match to a known address
var ErrUnableToGeocode = errors.New("unable to geocode")

// ErrRequestFailed either the connection was lost or similar
var ErrRequestFailed = errors.New("HTTP request failed")

// SolidifiedReverse does two requests to the nominatim APIs one with zoom level 10 and one with
// zoom level 18 and combines the results from each of them. This will fix the
// 'city', 'town', 'village', 'hamlet', 'suburb' reported by OSM.
// See https://github.com/openstreetmap/Nominatim/issues/885 for details.
func (c *Client) SolidifiedReverse(ctx context.Context, lat, lon float64) (*Result, error) {
	wq := synctx.NewWorkQueue(ctx)
	var resultHigh, resultLow *Result

	// ignore errors for the zoom level 10, if there are any errors
	// for that zoom level the request will simply not be enriched
	wq.Add("query nominatim zoom level 10", func(ctx context.Context) error {
		resultHigh, _ = c.Reverse(ctx, lat, lon, 10) // nolint: errcheck,gosec
		return nil
	})

	wq.Add("query nominatim zoom level 18", func(ctx context.Context) (err error) {
		resultLow, err = c.Reverse(ctx, lat, lon, 18)
		return
	})

	wq.Wait()

	if err := wq.Err(); err != nil {
		return nil, err
	}

	if resultHigh != nil {
		if resultLow.Address.City == "" {
			resultLow.Address.City = resultHigh.Address.City
		}
		if resultLow.Address.Town == "" {
			resultLow.Address.Town = resultHigh.Address.Town
		}
		if resultLow.Address.Village == "" {
			resultLow.Address.Village = resultHigh.Address.Village
		}
		if resultLow.Address.Hamlet == "" {
			resultLow.Address.Hamlet = resultHigh.Address.Hamlet
		}
		if resultLow.Address.Suburb == "" {
			resultLow.Address.Suburb = resultHigh.Address.Suburb
		}
		if resultLow.Address.County == "" {
			resultLow.Address.County = resultHigh.Address.County
		}
		if resultLow.Address.StateDistrict == "" {
			resultLow.Address.StateDistrict = resultHigh.Address.StateDistrict
		}
		if resultLow.Address.State == "" {
			resultLow.Address.State = resultHigh.Address.State
		}
	}

	return resultLow, nil
}

// Reverse executes a request to the nominatim service and returns
// the genuine result
func (c *Client) Reverse(ctx context.Context, lat, lon float64, zoom int) (*Result, error) {
	var result Result
	var err error

	// add logging & tracing statement
	startTime := time.Now()
	span, _ := opentracing.StartSpanFromContext(ctx, "Nominatim reverse",
		opentracing.StartTime(startTime))
	defer func() {
		dur := float64(time.Since(startTime)) / float64(time.Millisecond)
		le := log.Ctx(ctx).Debug().
			Float64("lat", lat).
			Float64("lon", lon).
			Int("zoom", zoom).
			Float64("duration", dur)

		// add error or result set info
		if err != nil {
			le = le.Err(err)
			span.LogFields(olog.Error(err))
		} else {
			span.LogFields(olog.String("address", result.DisplayName))
			le = le.Str("address", result.DisplayName)
		}

		le.Msg("Nominatim reverse")
	}()
	span.LogFields(
		olog.Float64("lat", lat),
		olog.Float64("lon", lon),
		olog.Int("zoom", zoom))
	defer span.Finish()

	// prepate request
	u, err := url.Parse(c.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse nominatim endpoint URL %q: %v", c.Endpoint, err)
	}
	u.Path = "nominatim/reverse"
	values := make(url.Values)
	values.Add("format", "jsonv2")
	values.Add("lat", strconv.FormatFloat(lat, 'f', 10, 64))
	values.Add("lon", strconv.FormatFloat(lon, 'f', 10, 64))
	values.Add("zoom", strconv.Itoa(zoom))
	u.RawQuery = values.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// execute request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // nolint: errcheck

	if resp.StatusCode != http.StatusOK {
		err = ErrRequestFailed
		return nil, err
	}

	// parse response
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	// Handle geocoding error
	if result.Error == ErrUnableToGeocode.Error() {
		err = ErrUnableToGeocode
		return nil, err
	}

	return &result, nil
}
