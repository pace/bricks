// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/10/30 by Vincent Landgraf

package client

import (
	"context"
	"net/url"
	"reflect"
	"testing"

	"github.com/rs/zerolog/log"
)

func TestClientURL(t *testing.T) {
	type fields struct {
		Endpoint string
		Language string
	}
	type args struct {
		path   string
		values url.Values
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "Endpoint invalid invalid",
			fields:  fields{"http://[fe80::%31%25en0]:8080/", "de"},
			args:    args{},
			want:    "",
			wantErr: true,
		},
		{
			name:    "Valid URL and generated path",
			fields:  fields{"https://j-1-dev.pacelink.net", "de"},
			args:    args{"/api/v1/vehicles/ID", nil},
			want:    "https://j-1-dev.pacelink.net/api/v1/vehicles/ID",
			wantErr: false,
		},
		{
			name:   "Valid URL and generated path with values",
			fields: fields{"https://j-1-dev.pacelink.net", "de"},
			args: args{"/api/v1/vehicles/ID", url.Values{
				"a": []string{":b"},
			}},
			want:    "https://j-1-dev.pacelink.net/api/v1/vehicles/ID?a=%3Ab",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Endpoint: tt.fields.Endpoint,
				Language: tt.fields.Language,
			}
			got, err := c.URL(tt.args.path, tt.args.values)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.URL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil {
				if !reflect.DeepEqual(got.String(), tt.want) {
					t.Errorf("Client.URL() = %v, want %v", got.String(), tt.want)
				}
			}
		})
	}
}

func TestIntegrationGetJSON(t *testing.T) {
	if testing.Short() {
		return
	}
	ctx := log.With().Logger().WithContext(context.Background())

	c := New("https://j-1-stage.pacelink.net")

	t.Run("Success", func(t *testing.T) {

		u, err := c.URL("/api/v1/vehicles/b636483a-df57-49a8-ad48-57bd56f060e0", nil)
		if err != nil {
			t.Error(err)
		}

		var resp struct {
			V struct {
				ID int `json:"id"`
			} `json:"vehicle"`
		}

		err = c.GetJSON(ctx, u, &resp)
		if err != nil {
			t.Error(err)
		}

		if resp.V.ID != 42902 {
			t.Errorf("expected 42902 got: %d", resp.V.ID)
		}
	})

	t.Run("Failure", func(t *testing.T) {
		u, err := c.URL("/api/v1/vehicles/foobar", nil)
		if err != nil {
			t.Error(err)
		}

		var resp struct {
			V struct {
				ID int `json:"id"`
			} `json:"vehicle"`
		}

		err = c.GetJSON(ctx, u, &resp)
		if err != ErrNotFound {
			t.Errorf("expected ErrNotFound got: %v", err)
		}
	})
}
