// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/10/30 by Vincent Landgraf

package cloudwebv2

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// TripStatus a trip has a status, reflecting its processing state
type TripStatus int

const (
	// Started trip has started and is still ongoing
	Started TripStatus = iota
	// Finished trip has ended; data processing pending
	Finished
	// Processing trip data is being processed; displayed data still may change
	Processing
	// Ready all data has been processed
	Ready
)

const tripStatusName = "startedfinishedprocessingready"

var tripStatusIndex = [...]uint8{0, 7, 15, 25, 30}

var (
	tripStatusNameToValue = map[string]TripStatus{
		"started":    Started,
		"finished":   Finished,
		"processing": Processing,
		"ready":      Ready,
	}

	tripStatusValueToName = map[TripStatus]string{
		Started:    "started",
		Finished:   "finished",
		Processing: "processing",
		Ready:      "ready",
	}
)

func init() {
	var v TripStatus
	if _, ok := interface{}(v).(fmt.Stringer); ok {
		tripStatusNameToValue = map[string]TripStatus{
			interface{}(Started).(fmt.Stringer).String():    Started,
			interface{}(Finished).(fmt.Stringer).String():   Finished,
			interface{}(Processing).(fmt.Stringer).String(): Processing,
			interface{}(Ready).(fmt.Stringer).String():      Ready,
		}
	}
}

// MarshalJSON is generated so TripStatus satisfies json.Marshaler.
func (r TripStatus) MarshalJSON() ([]byte, error) {
	if s, ok := interface{}(r).(fmt.Stringer); ok {
		return json.Marshal(s.String())
	}
	s, ok := tripStatusValueToName[r]
	if !ok {
		return nil, fmt.Errorf("invalid TripStatus: %d", r)
	}
	return json.Marshal(s)
}

// UnmarshalJSON is generated so TripStatus satisfies json.Unmarshaler.
func (r *TripStatus) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("TripStatus should be a string, got %s", data)
	}
	v, ok := tripStatusNameToValue[s]
	if !ok {
		return fmt.Errorf("invalid TripStatus %q", s)
	}
	*r = v
	return nil
}

func (r TripStatus) String() string {
	if r < 0 || r >= TripStatus(len(tripStatusIndex)-1) {
		return "TripStatus(" + strconv.FormatInt(int64(r), 10) + ")"
	}
	return tripStatusName[tripStatusIndex[r]:tripStatusIndex[r+1]]
}
