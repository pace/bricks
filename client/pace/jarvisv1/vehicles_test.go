// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/10/30 by Vincent Landgraf

package jarvisv1

import (
	"context"
	"testing"

	"lab.jamit.de/pace/go-microservice/client/pace/client"

	"github.com/rs/zerolog/log"
)

func TestIntegrationGetVehicleModel(t *testing.T) {
	if testing.Short() {
		return
	}
	ctx := log.With().Logger().WithContext(context.Background())

	c := New(EndpointStaging)

	t.Run("Success", func(t *testing.T) {
		resp, err := c.GetVehicleModel(ctx, &GetVehicleModelRequest{ID: "b636483a-df57-49a8-ad48-57bd56f060e0"})
		if err != nil {
			t.Error(err)
		}
		if resp.Vehicle.ID != 42902 {
			t.Errorf("Client.GetVehicleModel() Vehicle.ID = %v, want %v", resp.Vehicle.ID, 42902)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := c.GetVehicleModel(ctx, &GetVehicleModelRequest{ID: "foobar"})
		if err != client.ErrNotFound {
			t.Error(err)
		}
	})
}
