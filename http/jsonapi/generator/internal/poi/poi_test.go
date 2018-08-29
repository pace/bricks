// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/29 by Vincent Landgraf

package poi

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
	"strconv"
	"testing"

	"lab.jamit.de/pace/web/libs/go-microservice/http/jsonapi/runtime"
)

type testService struct {
	t *testing.T
}

func (s *testService) GetCheckForPaceApp(ctx context.Context, w GetCheckForPaceAppResponseWriter, r *GetCheckForPaceAppRequest) error {
	if r.ParamLatitude != 41.859194 {
		s.t.Errorf("expected ParamLatitude to be %f, got: %f", r.ParamLatitude, 41.859194)
	}
	if r.ParamLongitude != -87.646984 {
		s.t.Errorf("expected ParamLongitude to be %f, got: %f", r.ParamLatitude, -87.646984)
	}
	if r.ParamAppType != "fueling" {
		s.t.Errorf("expected ParamAppType to be %q, got: %q", "fueling", r.ParamAppType)
	}
	if r.ParamGpsSource != "raw" {
		s.t.Errorf("expected ParamGpsSource to be %q, got: %q", "raw", r.ParamGpsSource)
	}

	appsResp := make(LocationBasedAppsResponse, 10)
	for i := 0; i < 10; i++ {
		appsResp[i] = &LocationBasedAppsResponseItem{}
		appsResp[i].ID = strconv.Itoa(i)
		appsResp[i].AndroidInstantAppURL = "https://foobar.com"
		appsResp[i].Title = "Some app"
	}

	w.OK(appsResp)

	return nil
}

func (s *testService) GetSearch(ctx context.Context, w GetSearchResponseWriter, r *GetSearchRequest) error {
	return nil
}

func TestHandler(t *testing.T) {
	r := Router(&testService{t})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/poi/beta/check-for-pace-app?"+
		"latitude=41.859194&longitude=-87.646984&appType=fueling&gpsSource=raw", nil)
	req.Header.Set("Accept", runtime.JSONAPIContentType)
	req.Header.Set("Content-Type", runtime.JSONAPIContentType)

	r.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected OK got: %d", resp.StatusCode)
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		t.Error(string(b[:]))
		return
	}

	var data struct {
		Data []map[string]interface{} `json:"data"`
	}
	err := json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Data) != 10 {
		t.Error("Expected 10 apps")
	}
	if data.Data[0]["type"] != "locationBasedApp" {
		t.Error("Expected type locationBasedApp")
	}
}
