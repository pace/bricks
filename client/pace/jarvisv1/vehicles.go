// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/10/30 by Vincent Landgraf

package jarvisv1

import (
	"context"
	"net/url"
	"strconv"
)

// GetVehicleModel fetches the vehicle model based on the passed request
func (c *Client) GetVehicleModel(ctx context.Context, r *GetVehicleModelRequest) (*GetVehicleModelResponse, error) {
	var q url.Values

	// only add build year if not 0
	if r.BuildYear != 0 {
		q = url.Values{"build_year": []string{strconv.Itoa(r.BuildYear)}}
	}

	u, err := c.URL("/api/v1/vehicles/"+r.ID, q)
	if err != nil {
		return nil, err
	}

	var resp GetVehicleModelResponse
	err = c.GetJSON(ctx, u, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetVehicleModelRequest ID is mandatory, BuildYear is optional (if it is 0 it will be ignored)
type GetVehicleModelRequest struct {
	ID        string // ID of car model
	BuildYear int    // build year to check compatibility for
}

// GetVehicleModelResponse successful API response data
type GetVehicleModelResponse struct {
	Vehicle *Vehicle `json:"vehicle"`
}

// Vehicle single model data
type Vehicle struct {
	ID                          int           `json:"id"`
	UUID                        string        `json:"uuid"`
	Hsn                         string        `json:"hsn"`
	Tsn                         string        `json:"tsn"`
	Tsn2                        string        `json:"tsn_2"`
	ManufacturerName            string        `json:"manufacturer_name"`
	Name                        string        `json:"name"`
	ModelRange                  string        `json:"model_range"`
	ModelIdentifier             string        `json:"model_identifier"`
	BodyType                    string        `json:"body_type"`
	FuelType                    string        `json:"fuel_type"`
	GearCount                   int           `json:"gear_count"`
	Torque                      string        `json:"torque"`
	PowerHp                     int           `json:"power_hp"`
	PowerKw                     int           `json:"power_kw"`
	BuildPeriodStart            int           `json:"build_period_start"`
	BuildPeriodEnd              int           `json:"build_period_end"`
	ModelRangePeriodStart       int           `json:"model_range_period_start"`
	ModelRangePeriodEnd         int           `json:"model_range_period_end"`
	Cylinders                   int           `json:"cylinders"`
	CubicCapacity               int           `json:"cubic_capacity"`
	Acceleration                float64       `json:"acceleration"`
	FuelConsumption             float64       `json:"fuel_consumption"`
	TopSpeed                    int           `json:"top_speed"`
	DoorCount                   int           `json:"door_count"`
	TankSize                    int           `json:"tank_size"`
	Height                      int           `json:"height"`
	Length                      int           `json:"length"`
	Width                       int           `json:"width"`
	WheelBase                   int           `json:"wheel_base"`
	Weight                      int           `json:"weight"`
	Compatibility               string        `json:"compatibility"`
	PossibleObdLocations        string        `json:"possible_obd_locations"`
	ObdNotes                    string        `json:"obd_notes"`
	CommandBlacklistFooBar      []interface{} `json:"command_blacklist: foo_bar"`
	UpdatedAt                   int           `json:"updated_at"`
	PollingIntervalInMs         int           `json:"polling_interval_in_ms"`
	PollingGapInMs              int           `json:"polling_gap_in_ms"`
	SleepTimeoutInS             int           `json:"sleep_timeout_in_s"`
	VoltageThreshold            int           `json:"voltage_threshold"`
	VoltageWakeUpEnabled        bool          `json:"voltage_wakeup_enabled"`
	VoltageSleepEnabled         bool          `json:"voltage_sleep_enabled"`
	ObdMileageCorrectionEnabled bool          `json:"obd_mileage_correction_enabled"`
}
