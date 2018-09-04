// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/28 by Vincent Landgraf

package runtime

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"reflect"
	"testing"
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

	val := struct {
		UUID string `valid:"uuid"`
	}{"cb855aff-f03c-4307-9a22-ab5fcc6b6d7c"}

	ok := ValidateRequest(rec, req, &val)

	if !ok {
		t.Error("expected to succeed with the validation")
	}
}
