// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/11/05 by Vincent Landgraf

package cloudwebv2

import (
	"context"
	"fmt"
	"net/url"
)

// GetTripGroup fetches the trips for a certain group identifier
func (c *Client) GetTripGroup(ctx context.Context, r *TripGroupRequest) (*TripGroupResponse, error) {
	u, err := c.URL(fmt.Sprintf("/api/web/v2/trip-groups/%s", r.GroupIdentifier), r.Query())
	if err != nil {
		return nil, err
	}

	var resp TripGroupResponse
	err = c.GetJSON(ctx, u, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// TripGroupRequest for trip groups with eco events
type TripGroupRequest struct {
	GroupIdentifier  string // UUID of the trip group
	UniqueIdentifier string // unique identifier of the user
}

// Query generates a HTTP query based on the request data
func (r *TripGroupRequest) Query() url.Values {
	q := make(url.Values)
	q.Set("unique_identifier", r.UniqueIdentifier)
	q.Set("embed", "geo,events")
	return q
}

// TripGroupResponse contains trips for a group and their eco events
type TripGroupResponse struct {
	Trips []*Trip `json:"trips,omitempty"`
}
