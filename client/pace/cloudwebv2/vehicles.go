// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/10/30 by Vincent Landgraf

package cloudwebv2

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// GetVehiclesStats returns various car stats calculated for ALL trips
// belonging to the currently authenticated user’s car.
func (c *Client) GetVehiclesStats(ctx context.Context, r *GetVehiclesStatsRequest) (*GetVehiclesStatsResponse, error) {
	u, err := c.URL(fmt.Sprintf("/api/web/v2/vehicles/%s/stats", r.Vin), r.Query())
	if err != nil {
		return nil, err
	}

	var resp GetVehiclesStatsResponse
	err = c.GetJSON(ctx, u, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetVehiclesStatsRequest for various car stats calculated for ALL trips in time frame.
//
// start_at and end_at may seem to be switched but we decided to use the go-back-in-time principle.
// This makes more sense when looking at the trips itself but we wanted this to be aligned with the cloud app api.
//
// Although start_at and end_at are both optional, they are mutually required, that is,
// if one is provided, the other must also. Otherwise the sole value sent will be ignored.
//
// This includes all trips with status ready. Ongoing trips or somewhy unfinished trips are not being counted.
type GetVehiclesStatsRequest struct {
	UniqueIdentifier string    // Unique identifier for this user
	Vin              string    // VIN of car to get stats for
	StartAt          time.Time // Needs to be younger than EndAt
	EndAt            time.Time // Needs to be older than StartAt
}

// Query generate the query based on the request data
func (r *GetVehiclesStatsRequest) Query() url.Values {
	q := make(url.Values)

	q.Set("unique_identifier", r.UniqueIdentifier)
	q.Set("vin", r.Vin)
	q.Set("start_at", strconv.FormatInt(r.StartAt.Unix(), 10))
	q.Set("end_at", strconv.FormatInt(r.EndAt.Unix(), 10))

	return q
}

// GetVehiclesStatsResponse for a single car
type GetVehiclesStatsResponse struct {
	CurrentMileage          int                 `json:"current_mileage"`
	TripCountTotal          int                 `json:"trip_count_total"`
	TripCountBusiness       int                 `json:"trip_count_business"`
	TripCountPersonal       int                 `json:"trip_count_personal"`
	TripCountWork           int                 `json:"trip_count_work"`
	RefuelCostsInCents      int                 `json:"refuel_costs_in_cents"`
	AvgSpeedInKmPerH        float64             `json:"avg_speed_in_km_per_h"`
	MaxSpeedInKmPerH        float64             `json:"max_speed_in_km_per_h"`
	AvgDistanceInKm         float64             `json:"avg_distance_in_km"`
	MaxDistanceInKm         float64             `json:"max_distance_in_km"`
	TotalDistanceInKm       float64             `json:"total_distance_in_km"`
	AvgEcoScore             float64             `json:"avg_eco_score"`
	AvgFuelUsagePer100Km    float64             `json:"avg_fuel_usage_per_100_km"`
	AvgCostsInCentsPer100Km int                 `json:"avg_costs_in_cents_per_100_km"`
	AvgDurationInS          int                 `json:"avg_duration_in_s"`
	MaxDurationInS          int                 `json:"max_duration_in_s"`
	KilometersDriven        []*KilometersDriven `json:"kilometers_driven"`
	EcoEventCounts          *EcoEventCounts     `json:"eco_event_counts"`
	TroubleCodesCount       int                 `json:"trouble_codes_count"`
	FirstRecordedMileageInM int                 `json:"first_recorded_mileage_in_m"`
	AvgMonthlyDistanceInM   int                 `json:"avg_monthly_distance_in_m"`
	CurrentMileageInM       int                 `json:"current_mileage_in_m"`
}

// KilometersDriven for a single car
type KilometersDriven struct {
	Date     string `json:"date"`
	Business int    `json:"business"`
	Personal int    `json:"personal"`
	Work     int    `json:"work"`
	Unset    int    `json:"unset"`
}

// EcoEventCounts for a single car
type EcoEventCounts struct {
	Rpm          int `json:"rpm"`
	Acceleration int `json:"acceleration"`
	Braking      int `json:"braking"`
	Idle         int `json:"idle"`
	Speeding     int `json:"speeding"`
}
