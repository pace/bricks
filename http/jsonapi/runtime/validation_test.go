// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

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

	expected := map[string]any{
		"errors": []any{
			map[string]any{
				"title":  "UUID is invalid",
				"detail": "foo does not validate as uuid",
				"status": "422",
				"source": map[string]any{
					"parameter": "/uuid",
				},
			},
			map[string]any{
				"title":  "Token is invalid",
				"detail": "bar does not validate as uuid",
				"status": "422",
				"source": map[string]any{
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
	req := httptest.NewRequest(http.MethodPost, "/", nil)

	ok := ValidateParameters(rec, req, &val)
	if ok {
		t.Error("expected to fail the validation")
	}

	resp := rec.Result()

	defer func() {
		err := resp.Body.Close()
		assert.NoError(t, err)
	}()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Error("expected UnprocessableEntity")
	}

	var data map[string]any

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, data) {
		fmt.Printf("expected %#v got: %#v", expected, data)
	}
}

func TestValidateRequest(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)

	type args struct {
		w    http.ResponseWriter
		r    *http.Request
		data any
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
