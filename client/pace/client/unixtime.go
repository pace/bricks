// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/10/30 by Vincent Landgraf

package client

import (
	"encoding/json"
	"fmt"
	"time"
)

type UnixTime time.Time

// MarshalJSON is generated so TripStatus satisfies json.Marshaler.
func (r UnixTime) MarshalJSON() ([]byte, error) {
	v := time.Time(r).Unix()
	return json.Marshal(v)
}

// UnmarshalJSON is generated so TripStatus satisfies json.Unmarshaler.
func (r *UnixTime) UnmarshalJSON(data []byte) error {
	var timestamp int64
	if err := json.Unmarshal(data, &timestamp); err != nil {
		return fmt.Errorf("Unix timestamp should be an number, got %s", data)
	}
	*r = UnixTime(time.Unix(timestamp, 0))
	return nil
}
