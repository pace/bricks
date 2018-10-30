// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/10/30 by Vincent Landgraf

package cloudwebv2

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"lab.jamit.de/pace/go-microservice/client/pace/client"
)

// ListAllTrips and pushes them into the passed trips channel. The first trip (oldest trip) will
// be pushed first the last trip (newest) will be pushed last.
func (c *Client) ListAllTrips(ctx context.Context, r *ListTripsRequest, trips chan<- *Trip) error {
	if r.Limit == 0 { // set a default limit on the requests
		r.Limit = 25
	}

	r.Order = OrderDESC

	if r.StartAt.IsZero() {
		r.StartAt = time.Now().UTC()
	}

	if r.EndAt.IsZero() {
		r.EndAt = time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)
	}

next:
	resp, err := c.ListTrips(ctx, r)
	// todo handle error
	if err != nil {
		return err
	}

	for _, trip := range resp.Trips {
		trips <- trip
	}

	if len(resp.Trips) > 0 {
		r.ContinueAt = time.Time(resp.Trips[len(resp.Trips)-1].EndTime)
		goto next
	}

	return nil
}

// ListTrips returns a list of trips for the current user and the given VIN
func (c *Client) ListTrips(ctx context.Context, r *ListTripsRequest) (*ListTripsResponse, error) {
	u, err := c.URL(fmt.Sprintf("/api/web/v2/vehicles/%s/trips", r.Vin), r.Query())
	if err != nil {
		return nil, err
	}

	var resp ListTripsResponse
	err = c.GetJSON(ctx, u, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ListTripsRequest request for ListTrips
type ListTripsRequest struct {
	Vin              string    // vehicle identification number
	UniqueIdentifier string    // unique identifier of the user
	StartAt          time.Time //	The point in time at which to scope trips to, going into the past
	EndAt            time.Time //	scope trips up to this point in time. When passing end_at you’ll get back all trips between start_at and end_at.
	ContinueAt       time.Time // Time at which to continue on from for pagination
	Attributes       []string  // Filter the list of attributes you want to get back for each trip. For instance, if you specify attributes=vin,uuid, you will only get the vin and the uuid (applies to all trips in the collection).
	TripType         TripType  // Filter by trip type
	Limit            int       // The exact number of trips which the cloud will return.
	Mode             ListTripsRequestMode
	// Order define the order in which trips will be returned.
	// NOTE: If you sort by ASC the cloud will respond with the oldest trips first. Also sorting defaults to DESC when not otherwise specified therefore sending DESC is redundant.
	// Returns a 400 error if the value is neither of these two.
	Order Order
}

// Query generates a HTTP query based on the request data
func (r *ListTripsRequest) Query() url.Values {
	q := make(url.Values)

	q.Set("unique_identifier", r.UniqueIdentifier)

	if !r.StartAt.IsZero() {
		q.Set("start_at", strconv.FormatInt(r.StartAt.Unix(), 10))
	}
	if !r.EndAt.IsZero() {
		q.Set("end_at", strconv.FormatInt(r.EndAt.Unix(), 10))
	}
	if !r.ContinueAt.IsZero() {
		q.Set("continue_at", strconv.FormatInt(r.ContinueAt.Unix(), 10))
	}

	if r.Mode == "" {
		q.Set("mode", string(ModeSinglePage))
	} else {
		q.Set("mode", string(r.Mode))
	}

	if r.TripType != TripTypeUnset {
		q.Set("trip_type", string(r.TripType))
	}

	if r.Order == "" { // default order ASC
		q.Set("order", string(OrderASC))
	} else {
		q.Set("order", string(r.Order))
	}

	if r.Limit != 0 {
		q.Set("limit", strconv.Itoa(r.Limit))
	}

	if len(r.Attributes) > 0 {
		q.Set("attributes", strings.Join(r.Attributes, ","))
	}

	return q
}

// ListTripsResponse response for ListTrips
type ListTripsResponse struct {
	Trips []*Trip `json:"trips"`
}

// TripType accepts the values business, work, personal, unset
type TripType string

const (
	// TripTypeUnset type of the trip is not set by the user
	TripTypeUnset = ""
	// TripTypeBusiness business trip
	TripTypeBusiness = "business"
	// TripTypeWork work trip
	TripTypeWork = "work"
	// TripTypePersonal personal trip
	TripTypePersonal = "personal"
)

// Order accepts either ASC or DESC
type Order string

const (
	// OrderASC order asc
	OrderASC Order = "ASC"
	// OrderDESC order desc
	OrderDESC Order = "DESC"
)

// ListTripsRequestMode	accepts all_pages or single_page, defaults to single_page
type ListTripsRequestMode string

const (
	// ModeAllPages returns all of the pages upto the continue at plus 25 trips
	ModeAllPages ListTripsRequestMode = "all_pages"
	// ModeSinglePage returns the next 25 trips from continue_at onwards.
	ModeSinglePage ListTripsRequestMode = "single_page"
)

// Trip individual trip of a user
type Trip struct {
	UUID                string          `json:"uuid"`
	Vin                 string          `json:"vin"`
	StartMileage        int             `json:"start_mileage"`
	StartTime           client.UnixTime `json:"start_time"`
	EndMileage          int             `json:"end_mileage"`
	EndTime             client.UnixTime `json:"end_time"`
	GroupIdentifier     string          `json:"group_identifier"`
	Status              TripStatus      `json:"status"`
	TripType            string          `json:"trip_type"`
	DriverName          string          `json:"driver_name"`
	StartAddress        *TripAddress    `json:"start_address"`
	EndAddress          *TripAddress    `json:"end_address"`
	StartPosition       *TripPosition   `json:"start_position"`
	EndPosition         *TripPosition   `json:"end_position"`
	BusinessReason      string          `json:"business_reason"`
	BusinessPartnerName string          `json:"business_partner_name"`
	Stats               *TripStats      `json:"stats"`
	EcoScores           *TripEcoScores  `json:"eco_scores"`
	SafetyScore         int             `json:"safety_score"`
	ManuallyCreated     bool            `json:"manually_created"`
	UpdatedAt           int             `json:"updated_at"`
}

// TripEcoScores of a trip
type TripEcoScores struct {
	RpmScore          float64 `json:"rpm_score"`
	AccelerationScore float64 `json:"acceleration_score"`
	BrakingScore      float64 `json:"braking_score"`
	IdleScore         float64 `json:"idle_score"`
	SpeedingScore     float64 `json:"speeding_score"`
	TotalScore        float64 `json:"total_score"`
}

// TripStats of a trip
type TripStats struct {
	AvgSpeedInKmPerH     float64 `json:"avg_speed_in_km_per_h"`
	AvgRpm               float64 `json:"avg_rpm"`
	MaxSpeedInKmPerH     float64 `json:"max_speed_in_km_per_h"`
	MaxRpm               int     `json:"max_rpm"`
	DistanceInKm         float64 `json:"distance_in_km"`
	CostInCents          int     `json:"cost_in_cents"`
	DurationInS          int     `json:"duration_in_s"`
	AvgFuelUsagePer100Km float64 `json:"avg_fuel_usage_per_100_km"`
}

// TripAddress of trip start and end address
type TripAddress struct {
	Name             string `json:"name"`
	SourceIdentifier string `json:"source_identifier"`
	Street           string `json:"street"`
	HouseNumber      string `json:"house_number"`
	Zip              string `json:"zip"`
	Suburb           string `json:"suburb"`
	State            string `json:"state"`
	City             string `json:"city"`
	Country          string `json:"country"`
	AddressType      string `json:"address_type"`
}

// TripPosition of trip start and end position
type TripPosition struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
}
