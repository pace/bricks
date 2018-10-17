// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/10/12 by Vincent Landgraf

package nominatim

import (
	"context"
	"testing"

	"github.com/rs/zerolog/log"
)

func TestIntegrationReverse(t *testing.T) {
	ctx := context.Background()
	ctx = log.With().Logger().WithContext(ctx)

	_, err := DefaultClient.Reverse(ctx, 0, 0, 18)
	if err != ErrUnableToGeocode {
		t.Error("expected unable to geocode error, got: ", err)
	}

	res, err := DefaultClient.Reverse(ctx, 49.01251, 8.42636, 18)
	if err != nil {
		t.Error("expected error, got: ", err)
	}

	expected := "Haid-und-Neu-Straße 18, 76131 Karlsruhe"
	if res.Address.GermanShort() != expected {
		t.Errorf("expected %q, got: %q", expected, res.Address.GermanShort())
	}
}

func TestIntegrationSolidifiedReverse(t *testing.T) {
	ctx := context.Background()
	ctx = log.With().Logger().WithContext(ctx)

	tests := []struct {
		name    string
		lat     float64
		long    float64
		wantErr bool
	}{
		{"Herrenstraße 31, 76133 Karlsruhe", 49.00825, 8.39765, false},
		{"Strange Place", 0, 0, true},
		{"Hauptstraße 5, 71732 Tamm", 48.91943465, 9.11269735, false},
		{"Haid-und-Neu-Straße 18, 76131 Karlsruhe", 49.01251, 8.42636, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DefaultClient.SolidifiedReverse(ctx, tt.lat, tt.long)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.SolidifiedReverse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got.Address.GermanShort() != tt.name {
				t.Errorf("Client.SolidifiedReverse() = %v, want %v", got.Address.GermanShort(), tt.name)
			}
		})
	}
}
