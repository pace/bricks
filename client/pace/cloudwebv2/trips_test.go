// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/10/30 by Vincent Landgraf

package cloudwebv2

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
)

func TestListAllTrips(t *testing.T) {
	ctx := log.With().Logger().WithContext(context.Background())
	c := New(EndpointProduction)
	trips := make(chan *Trip)
	go func() {
		err := c.ListAllTrips(ctx, &ListTripsRequest{
			UniqueIdentifier: "14dff3a0-93ce-4f19-8c54-93adf720eea7",
			Vin:              "WVWZZZAUZFP613749",
		}, trips)
		if err != nil {
			t.Error(err)
		}
		close(trips)
	}()

	i := 0
	km := 0.0
	for trip := range trips {
		t.Logf("Trip: %s", trip.GroupIdentifier)
		i++
		km += trip.Stats.DistanceInKm
	}
	t.Logf("Trips: %d %f", i, km)
}

func TestListTripsRequest_Query(t *testing.T) {
	type fields struct {
		Vin              string
		UniqueIdentifier string
		StartAt          time.Time
		EndAt            time.Time
		ContinueAt       time.Time
		Attributes       []string
		TripType         TripType
		Limit            int
		Mode             ListTripsRequestMode
		Order            Order
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Example query",
			fields: fields{
				UniqueIdentifier: "b822c0ed-bc11-494e-8e24-f7e683611eeb",
				Mode:             ModeAllPages,
			},
			want: "mode=all_pages&order=ASC&unique_identifier=b822c0ed-bc11-494e-8e24-f7e683611eeb",
		},
		{
			name: "With trip type order and limit query",
			fields: fields{
				UniqueIdentifier: "b822c0ed-bc11-494e-8e24-f7e683611eeb",
				TripType:         TripTypePersonal,
				Order:            OrderDESC,
				Limit:            25,
			},
			want: "limit=25&mode=single_page&order=DESC&trip_type=personal&unique_identifier=b822c0ed-bc11-494e-8e24-f7e683611eeb",
		},
		{
			name: "With attributes",
			fields: fields{
				UniqueIdentifier: "b822c0ed-bc11-494e-8e24-f7e683611eeb",
				Attributes:       []string{"vin", "uuid"},
			},
			want: "attributes=vin%2Cuuid&mode=single_page&order=ASC&unique_identifier=b822c0ed-bc11-494e-8e24-f7e683611eeb",
		},
		{
			name: "With time",
			fields: fields{
				UniqueIdentifier: "b822c0ed-bc11-494e-8e24-f7e683611eeb",
				StartAt:          time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC),
				EndAt:            time.Date(2010, 11, 17, 20, 34, 58, 651387237, time.UTC),
				ContinueAt:       time.Date(2009, 5, 17, 20, 34, 58, 651387237, time.UTC),
			},
			want: "continue_at=1242592498&end_at=1290026098&mode=single_page&order=ASC&start_at=1258490098&unique_identifier=b822c0ed-bc11-494e-8e24-f7e683611eeb",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ListTripsRequest{
				Vin:              tt.fields.Vin,
				UniqueIdentifier: tt.fields.UniqueIdentifier,
				StartAt:          tt.fields.StartAt,
				EndAt:            tt.fields.EndAt,
				ContinueAt:       tt.fields.ContinueAt,
				Attributes:       tt.fields.Attributes,
				TripType:         tt.fields.TripType,
				Limit:            tt.fields.Limit,
				Mode:             tt.fields.Mode,
				Order:            tt.fields.Order,
			}
			if got := r.Query().Encode(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListTripsRequest.Query() = %v, want %v", got, tt.want)
			}
		})
	}
}
