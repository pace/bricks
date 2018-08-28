// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/24 by Vincent Landgraf

package runtime

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestScanNumericParametersInPath(t *testing.T) {
	req := httptest.NewRequest("GET", "/foo/", nil)
	rec := httptest.NewRecorder()
	var param0 uint
	var param1 uint8
	var param2 uint16
	var param3 uint32
	var param4 uint64
	var param10 int
	var param11 int8
	var param12 int16
	var param13 int32
	var param14 int64
	var param20 float32
	var param21 float64
	ok := ScanParameters(rec, req,
		&ScanParameter{&param0, ScanInPath, "12"},
		&ScanParameter{&param1, ScanInPath, "12"},
		&ScanParameter{&param2, ScanInPath, "12"},
		&ScanParameter{&param3, ScanInPath, "12"},
		&ScanParameter{&param4, ScanInPath, "12"},
		&ScanParameter{&param10, ScanInPath, "-12"},
		&ScanParameter{&param11, ScanInPath, "-12"},
		&ScanParameter{&param12, ScanInPath, "-12"},
		&ScanParameter{&param13, ScanInPath, "-12"},
		&ScanParameter{&param14, ScanInPath, "-12"},
		&ScanParameter{&param20, ScanInPath, "-12.123123123123123123123123"},
		&ScanParameter{&param21, ScanInPath, "-12.123123123123123123123123"},
	)

	// Parsing
	if !ok {
		t.Errorf("expected the scanning to be successful")
	}

	// Uint
	if param0 != uint(12) {
		t.Errorf("expected parsing result %#v got: %#v", uint(12), param0)
	}
	if param1 != uint8(12) {
		t.Errorf("expected parsing result %#v got: %#v", uint8(12), param1)
	}
	if param2 != uint16(12) {
		t.Errorf("expected parsing result %#v got: %#v", uint16(12), param2)
	}
	if param3 != uint32(12) {
		t.Errorf("expected parsing result %#v got: %#v", uint32(12), param3)
	}
	if param4 != uint64(12) {
		t.Errorf("expected parsing result %#v got: %#v", uint64(12), param4)
	}

	// Int
	if param10 != int(-12) {
		t.Errorf("expected parsing result %#v got: %#v", int(-12), param10)
	}
	if param11 != int8(-12) {
		t.Errorf("expected parsing result %#v got: %#v", int8(-12), param11)
	}
	if param12 != int16(-12) {
		t.Errorf("expected parsing result %#v got: %#v", int16(-12), param12)
	}
	if param13 != int32(-12) {
		t.Errorf("expected parsing result %#v got: %#v", int32(-12), param13)
	}
	if param14 != int64(-12) {
		t.Errorf("expected parsing result %#v got: %#v", int64(-12), param14)
	}

	// Float
	if param20 != float32(-12.123123123123123123123123) {
		t.Errorf("expected parsing result %#v got: %#v", float32(-12.123123123123123123123123), param20)
	}
	if param21 != float64(-12.123123123123123123123123) {
		t.Errorf("expected parsing result %#v got: %#v", float64(-12.123123123123123123123123), param21)
	}
}

func TestScanNumericParametersInQueryUint(t *testing.T) {
	req := httptest.NewRequest("GET", "/foo?num=12", nil)
	rec := httptest.NewRecorder()
	var param0 uint
	var param1 uint8
	var param2 uint16
	var param3 uint32
	var param4 uint64
	ok := ScanParameters(rec, req,
		&ScanParameter{&param0, ScanInQuery, "num"},
		&ScanParameter{&param1, ScanInQuery, "num"},
		&ScanParameter{&param2, ScanInQuery, "num"},
		&ScanParameter{&param3, ScanInQuery, "num"},
		&ScanParameter{&param4, ScanInQuery, "num"},
	)

	// Parsing
	if !ok {
		t.Errorf("expected the scanning to be successful")
	}

	// Uint
	if param0 != uint(12) {
		t.Errorf("expected parsing result %#v got: %#v", uint(12), param0)
	}
	if param1 != uint8(12) {
		t.Errorf("expected parsing result %#v got: %#v", uint8(12), param1)
	}
	if param2 != uint16(12) {
		t.Errorf("expected parsing result %#v got: %#v", uint16(12), param2)
	}
	if param3 != uint32(12) {
		t.Errorf("expected parsing result %#v got: %#v", uint32(12), param3)
	}
	if param4 != uint64(12) {
		t.Errorf("expected parsing result %#v got: %#v", uint64(12), param4)
	}
}

func TestScanNumericParametersInQueryInt(t *testing.T) {
	req := httptest.NewRequest("GET", "/foo?num=-12", nil)
	rec := httptest.NewRecorder()
	var param10 int
	var param11 int8
	var param12 int16
	var param13 int32
	var param14 int64
	ok := ScanParameters(rec, req,
		&ScanParameter{&param10, ScanInQuery, "num"},
		&ScanParameter{&param11, ScanInQuery, "num"},
		&ScanParameter{&param12, ScanInQuery, "num"},
		&ScanParameter{&param13, ScanInQuery, "num"},
		&ScanParameter{&param14, ScanInQuery, "num"},
	)

	// Parsing
	if !ok {
		t.Errorf("expected the scanning to be successful")
	}

	// Iint
	if param10 != int(-12) {
		t.Errorf("expected parsing result %#v got: %#v", int(-12), param10)
	}
	if param11 != int8(-12) {
		t.Errorf("expected parsing result %#v got: %#v", int8(-12), param11)
	}
	if param12 != int16(-12) {
		t.Errorf("expected parsing result %#v got: %#v", int16(-12), param12)
	}
	if param13 != int32(-12) {
		t.Errorf("expected parsing result %#v got: %#v", int32(-12), param13)
	}
	if param14 != int64(-12) {
		t.Errorf("expected parsing result %#v got: %#v", int64(-12), param14)
	}
}

func TestScanNumericParametersInQueryFloat(t *testing.T) {
	req := httptest.NewRequest("GET", "/foo?num=-12.123123123123123123123123", nil)
	rec := httptest.NewRecorder()
	var param20 float32
	var param21 float64
	ok := ScanParameters(rec, req,
		&ScanParameter{&param20, ScanInQuery, "num"},
		&ScanParameter{&param21, ScanInQuery, "num"},
	)

	// Parsing
	if !ok {
		t.Errorf("expected the scanning to be successful")
	}

	// Float
	if param20 != float32(-12.123123123123123123123123) {
		t.Errorf("expected parsing result %#v got: %#v", float32(-12.123123123123123123123123), param20)
	}
	if param21 != float64(-12.123123123123123123123123) {
		t.Errorf("expected parsing result %#v got: %#v", float64(-12.123123123123123123123123), param21)
	}
}

func TestScanNumericParametersInQueryFloatArray(t *testing.T) {
	req := httptest.NewRequest("GET", "/foo?num=-12.123123123123123123123123&num=-987.123123123123123123123123", nil)
	rec := httptest.NewRecorder()
	var param []float32
	ok := ScanParameters(rec, req,
		&ScanParameter{&param, ScanInQuery, "num"},
	)

	// Parsing
	if !ok {
		t.Errorf("expected the scanning to be successful")
	}

	if len(param) != 2 {
		t.Fatalf("expected the scanning to create 2 results got: %d", len(param))
	}

	// Float
	if param[0] != float32(-12.123123123123123123123123) {
		t.Errorf("expected parsing result %#v got: %#v", float32(-12.123123123123123123123123), param[0])
	}
	if param[1] != float32(-987.123123123123123123123123) {
		t.Errorf("expected parsing result %#v got: %#v", float32(-987.123123123123123123123123), param[1])
	}
}

func TestScanNumericParametersInQueryFloatArrayFail(t *testing.T) {
	req := httptest.NewRequest("GET", "/foo?num=-12.123123123123123123123123&num=stuff", nil)
	rec := httptest.NewRecorder()
	var param []float32
	ok := ScanParameters(rec, req,
		&ScanParameter{&param, ScanInQuery, "num"},
	)

	// Parsing
	if ok {
		t.Errorf("expected the scanning to be failing")
	}

	resp := rec.Result()
	defer resp.Body.Close()

	var errList errorObjects
	dec := json.NewDecoder(resp.Body)
	err := dec.Decode(&errList)
	if err != nil {
		t.Fatal(err)
	}

	if len(errList.List) != 1 {
		t.Fatal("there must be one error in the list, got none")
	}

	errObj := errList.List[0]
	if r := "invalid value, exepcted float32 got: \"stuff\""; errObj.Title != r {
		t.Errorf("expected title %q got: %q", r, errObj.Title)
	}
	if r := "400"; errObj.Status != r {
		t.Errorf("expected status %q got: %q", r, errObj.Status)
	}
	if r := "num"; (*errObj.Source)["parameter"] != r {
		t.Errorf("expected source parameter %q got: %q", r, (*errObj.Source)["parameter"])
	}
}

func TestScanParametersError(t *testing.T) {
	req := httptest.NewRequest("GET", "/foo?num=-12", nil)
	rec := httptest.NewRecorder()
	var param uint
	ok := ScanParameters(rec, req,
		&ScanParameter{&param, ScanInQuery, "num"},
	)

	// Parsing
	if ok {
		t.Errorf("expected the scanning to be failing")
	}

	resp := rec.Result()
	defer resp.Body.Close()

	var errList errorObjects
	dec := json.NewDecoder(resp.Body)
	err := dec.Decode(&errList)
	if err != nil {
		t.Fatal(err)
	}

	if len(errList.List) != 1 {
		t.Fatal("there must be one error in the list, got none")
	}

	errObj := errList.List[0]
	if r := "expected integer"; errObj.Title != r {
		t.Errorf("expected title %q got: %q", r, errObj.Title)
	}
	if r := "400"; errObj.Status != r {
		t.Errorf("expected status %q got: %q", r, errObj.Status)
	}
	if r := "num"; (*errObj.Source)["parameter"] != r {
		t.Errorf("expected source parameter %q got: %q", r, (*errObj.Source)["parameter"])
	}
}
