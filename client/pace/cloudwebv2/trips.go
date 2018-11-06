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

// ListTripsRequestMode accepts all_pages or single_page, defaults to single_page
type ListTripsRequestMode string

const (
	// ModeAllPages returns all of the pages up to the continue at plus 25 trips
	ModeAllPages ListTripsRequestMode = "all_pages"
	// ModeSinglePage returns the next 25 trips from continue_at onwards.
	ModeSinglePage ListTripsRequestMode = "single_page"
)

// Trip individual trip of a user
type Trip struct {
	ID                  int             `json:"id,omitempty"`
	UUID                string          `json:"uuid,omitempty"`
	Vin                 string          `json:"vin,omitempty"`
	StartMileage        int             `json:"start_mileage,omitempty"`
	StartMileageInM     int             `json:"start_mileage_in_m,omitempty"`
	StartTime           client.UnixTime `json:"start_time,omitempty"`
	EndMileage          int             `json:"end_mileage,omitempty"`
	EndMileageInM       int             `json:"end_mileage_in_m,omitempty"`
	EndTime             client.UnixTime `json:"end_time,omitempty"`
	GroupIdentifier     string          `json:"group_identifier,omitempty"`
	Status              TripStatus      `json:"status,omitempty"`
	TripType            string          `json:"trip_type,omitempty"`
	DriverName          string          `json:"driver_name,omitempty"`
	StartAddress        *TripAddress    `json:"start_address,omitempty"`
	EndAddress          *TripAddress    `json:"end_address,omitempty"`
	StartPosition       *TripPosition   `json:"start_position,omitempty"`
	EndPosition         *TripPosition   `json:"end_position,omitempty"`
	BusinessReason      string          `json:"business_reason,omitempty"`
	BusinessPartnerName string          `json:"business_partner_name,omitempty"`
	Stats               *TripStats      `json:"stats,omitempty"`
	EcoScores           *TripEcoScores  `json:"eco_scores,omitempty"`
	SafetyScore         float64         `json:"safety_score,omitempty"`
	ManuallyCreated     bool            `json:"manually_created,omitempty"`
	UpdatedAt           client.UnixTime `json:"updated_at,omitempty"`
	GeoScore            []*GeoScore     `json:"geo,omitempty"`
	Events              []*GeoEvent     `json:"events,omitempty"`
}

// TripEcoScores of a trip
type TripEcoScores struct {
	RpmScore          float64 `json:"rpm_score,omitempty"`
	AccelerationScore float64 `json:"acceleration_score,omitempty"`
	BrakingScore      float64 `json:"braking_score,omitempty"`
	IdleScore         float64 `json:"idle_score,omitempty"`
	SpeedingScore     float64 `json:"speeding_score,omitempty"`
	TotalScore        float64 `json:"total_score,omitempty"`
}

// TripStats of a trip
type TripStats struct {
	AvgSpeedInKmPerH     float64 `json:"avg_speed_in_km_per_h,omitempty"`
	AvgRpm               float64 `json:"avg_rpm,omitempty"`
	MaxSpeedInKmPerH     float64 `json:"max_speed_in_km_per_h,omitempty"`
	MaxRpm               float64 `json:"max_rpm,omitempty"`
	DistanceInKm         float64 `json:"distance_in_km,omitempty"`
	DistanceInM          int     `json:"distance_in_m"`
	CostInCents          float64 `json:"cost_in_cents,omitempty"`
	DurationInS          float64 `json:"duration_in_s,omitempty"`
	AvgFuelUsagePer100Km float64 `json:"avg_fuel_usage_per_100_km,omitempty"`
}

// TripAddress of trip start and end address
type TripAddress struct {
	Name             string `json:"name,omitempty"`
	SourceIdentifier string `json:"source_identifier,omitempty"`
	Street           string `json:"street,omitempty"`
	HouseNumber      string `json:"house_number,omitempty"`
	Zip              string `json:"zip,omitempty"`
	Suburb           string `json:"suburb,omitempty"`
	State            string `json:"state,omitempty"`
	City             string `json:"city,omitempty"`
	Country          string `json:"country,omitempty"`
	AddressType      string `json:"address_type,omitempty"`
}

// TripPosition of trip start and end position
type TripPosition struct {
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	Altitude  float64 `json:"altitude,omitempty"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
}
