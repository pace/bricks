// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/28 by Vincent Landgraf

package runtime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateParametersWithError(t *testing.T) {
	type access struct {
		Token string `valid:"uuid"`
	}
	type input struct {
		UUID   string `valid:"uuid"`
		Access access
	}
	expected := map[string]interface{}{
		"errors": []interface{}{
			map[string]interface{}{
				"title":  "UUID is invalid",
				"detail": "foo does not validate as uuid",
				"status": "422",
				"source": map[string]interface{}{
					"parameter": "/uuid",
				},
			},
			map[string]interface{}{
				"title":  "Token is invalid",
				"detail": "bar does not validate as uuid",
				"status": "422",
				"source": map[string]interface{}{
					"parameter": "/access/token",
				},
			},
		},
	}
	val := input{
		UUID: "foo",
		Access: access{
			"bar",
		},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/", nil)

	ok := ValidateParameters(rec, req, &val)

	if ok {
		t.Error("expected to fail the validation")
	}

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != 422 {
		t.Error("expected UnprocessableEntity")
	}

	var data map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, data) {
		fmt.Printf("expected %#v got: %#v", expected, data)
	}
}

func TestValidateRequest(t *testing.T) {

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/", nil)

	type args struct {
		w    http.ResponseWriter
		r    *http.Request
		data interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "uuid lowercase",
			args: args{
				w: rec,
				r: req,
				data: struct {
					UUID string `valid:"uuid"`
				}{"cb855aff-f03c-4307-9a22-ab5fcc6b6d7c"},
			},
			want: true,
		},
		{
			name: "uuid uppercase",
			args: args{
				w: rec,
				r: req,
				data: struct {
					UUID string `valid:"uuid"`
				}{"CB855AFF-F03C-4307-9A22-AB5FCC6B6D7C"},
			},
			want: true,
		},
		{
			name: "uuid mixed lower / uppercase",
			args: args{
				w: rec,
				r: req,
				data: struct {
					UUID string `valid:"uuid"`
				}{"CB855AFF-F03C-4307-9a22-ab5fcc6b6d7c"},
			},
			want: true,
		},
		{
			name: "invalid uuid",
			args: args{
				w: rec,
				r: req,
				data: struct {
					UUID string `valid:"uuid"`
				}{"hey-uuid"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, ValidateRequest(tt.args.w, tt.args.r, tt.args.data), "ValidateRequest(%v, %v, %v)", tt.args.w, tt.args.r, tt.args.data)
		})
	}
}
