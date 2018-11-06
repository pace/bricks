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
	u, err := c.URL("/api/v1/vehicles/"+r.ID, r.Query())
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

// Query generates a HTTP query based on the request data
func (r *GetVehicleModelRequest) Query() url.Values {
	var q url.Values

	// only add build year if not 0
	if r.BuildYear != 0 {
		q = url.Values{"build_year": []string{strconv.Itoa(r.BuildYear)}}
	}

	return q
}

// GetVehicleModelResponse successful API response data
type GetVehicleModelResponse struct {
	Vehicle *Vehicle `json:"vehicle"`
}

// Vehicle single model data
type Vehicle struct {
	ID                          int           `json:"id,omitempty"`
	UUID                        string        `json:"uuid,omitempty"`
	Hsn                         string        `json:"hsn,omitempty"`
	Tsn                         string        `json:"tsn,omitempty"`
	Tsn2                        string        `json:"tsn_2,omitempty"`
	ManufacturerName            string        `json:"manufacturer_name,omitempty"`
	Name                        string        `json:"name,omitempty"`
	ModelRange                  string        `json:"model_range,omitempty"`
	ModelIdentifier             string        `json:"model_identifier,omitempty"`
	BodyType                    string        `json:"body_type,omitempty"`
	FuelType                    string        `json:"fuel_type,omitempty"`
	GearCount                   int           `json:"gear_count,omitempty"`
	Torque                      string        `json:"torque,omitempty"`
	PowerHp                     int           `json:"power_hp,omitempty"`
	PowerKw                     int           `json:"power_kw,omitempty"`
	BuildPeriodStart            int           `json:"build_period_start,omitempty"`
	BuildPeriodEnd              int           `json:"build_period_end,omitempty"`
	ModelRangePeriodStart       int           `json:"model_range_period_start,omitempty"`
	ModelRangePeriodEnd         int           `json:"model_range_period_end,omitempty"`
	Cylinders                   int           `json:"cylinders,omitempty"`
	CubicCapacity               int           `json:"cubic_capacity,omitempty"`
	Acceleration                float64       `json:"acceleration,omitempty"`
	FuelConsumption             float64       `json:"fuel_consumption,omitempty"`
	TopSpeed                    int           `json:"top_speed,omitempty"`
	DoorCount                   int           `json:"door_count,omitempty"`
	TankSize                    int           `json:"tank_size,omitempty"`
	Height                      int           `json:"height,omitempty"`
	Length                      int           `json:"length,omitempty"`
	Width                       int           `json:"width,omitempty"`
	WheelBase                   int           `json:"wheel_base,omitempty"`
	Weight                      int           `json:"weight,omitempty"`
	Compatibility               string        `json:"compatibility,omitempty"`
	PossibleObdLocations        string        `json:"possible_obd_locations,omitempty"`
	ObdNotes                    string        `json:"obd_notes,omitempty"`
	CommandBlacklist            []interface{} `json:"command_blacklist,omitempty"`
	UpdatedAt                   int           `json:"updated_at,omitempty"`
	PollingIntervalInMs         int           `json:"polling_interval_in_ms,omitempty"`
	PollingGapInMs              int           `json:"polling_gap_in_ms,omitempty"`
	SleepTimeoutInS             int           `json:"sleep_timeout_in_s,omitempty"`
	VoltageThreshold            int           `json:"voltage_threshold,omitempty"`
	VoltageWakeUpEnabled        bool          `json:"voltage_wakeup_enabled,omitempty"`
	VoltageSleepEnabled         bool          `json:"voltage_sleep_enabled,omitempty"`
	ObdMileageCorrectionEnabled bool          `json:"obd_mileage_correction_enabled,omitempty"`
}
