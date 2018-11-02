// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/10/30 by Vincent Landgraf

package cloudwebv2

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gocarina/gocsv"
)

// GetTripGroupSamples returns the CVS or GPX data for the trip
func (c *Client) GetTripGroupSamples(ctx context.Context, r *GetTripGroupSamplesRequest) (*GetTripGroupSamplesResponse, error) {
	u, err := c.URL(fmt.Sprintf("/api/web/v2/trips/%s/samples", r.GroupIdentifier), r.Query())
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	switch r.Format {
	case TripSamplesCSV:
		req.Header.Set("Accept", "text/csv")
	case TripSamplesGPX:
		req.Header.Set("Accept", "application/gpx+xml")
	}

	resp, err := c.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close() // nolint: megacheck,errcheck

	var samples GetTripGroupSamplesResponse
	err = gocsv.Unmarshal(resp.Body, &samples.Rows)
	if err != nil {
		return nil, err
	}

	return &samples, nil
}

// GetTripGroupSamplesRequest contains the request parameters
type GetTripGroupSamplesRequest struct {
	UniqueIdentifier string
	GroupIdentifier  string // (required) group_identifer of the trips the trip samples are for.
	Format           TripSamplesFormat
}

// Query converts the request data into query parameters
func (r *GetTripGroupSamplesRequest) Query() url.Values {
	q := make(url.Values)

	q.Set("unique_identifier", r.UniqueIdentifier)

	return q
}

// GetTripGroupSamplesResponse contains the trip sample data
type GetTripGroupSamplesResponse struct {
	Rows []*TripSampleRow
}

// TripSamplesFormat format for the trip samples
type TripSamplesFormat int

const (
	// TripSamplesCSV Get all trip samples for a trip group in CSV  ¶
	TripSamplesCSV TripSamplesFormat = iota
	// TripSamplesGPX  GPX tracks of trip samples for a given trip.
	// This can be used to replay trips in XCode and will be used most
	// likely for debugging purposes of the mobile development teams.
	TripSamplesGPX
)

// TripSampleRow single trip data sample
type TripSampleRow struct {
	AcclPitch         float64 `csv:"Gyroscope Pitch" json:"gyroscope_pitch,omitempty"`
	AcclRoll          float64 `csv:"Gyroscope Roll" json:"gyroscope_roll,omitempty"`
	AcclYaw           float64 `csv:"Gyroscope Yaw" json:"gyroscope_roll_yaw,omitempty"`
	AcclX             float64 `csv:"Accelerometer X" json:"accelerometer_x,omitempty"`
	AcclY             float64 `csv:"Accelerometer Y" json:"accelerometer_y,omitempty"`
	AcclZ             float64 `csv:"Accelerometer Z" json:"accelerometer_z,omitempty"`
	ElapsedKilometers float64 `csv:"Elapsed distance" json:"elapsed_kilometers,omitempty"`
	ElapsedTime       float64 `csv:"Elapsed time" json:"elapsed_time,omitempty"`
	PosAlt            float64 `csv:"GPS Altitude" json:"pos_alt,omitempty"`
	PosLat            float64 `csv:"GPS Latitude" json:"pos_lat,omitempty"`
	PosLon            float64 `csv:"GPS Longitude" json:"pos_lon,omitempty"`
	Speed             float64 `csv:"Speed" json:"speed,omitempty"`
	FuelUsage         float64 `csv:"Consumed fuel" json:"fuel_usage,omitempty"`
	Rpm               int     `csv:"RPM" json:"rpm,omitempty"`
}
