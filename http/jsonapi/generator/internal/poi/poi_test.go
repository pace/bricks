// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/29 by Vincent Landgraf

package poi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pace/bricks/http/jsonapi/runtime"
	"github.com/pace/bricks/http/oauth2"
	oidc "github.com/pace/bricks/http/oidc"
	apikey "github.com/pace/bricks/http/security/apikey"
)

type testService struct {
	t *testing.T
}

func (s *testService) DeduplicatePoi(context.Context, DeduplicatePoiResponseWriter, *DeduplicatePoiRequest) error {
	return nil
}
func (s *testService) MovePoiAtPosition(context.Context, MovePoiAtPositionResponseWriter, *MovePoiAtPositionRequest) error {
	return nil
}
func (s *testService) CheckForPaceApp(ctx context.Context, w CheckForPaceAppResponseWriter, r *CheckForPaceAppRequest) error {
	if r.ParamFilterLatitude != 41.859194 {
		s.t.Errorf("expected ParamLatitude to be %f, got: %f", 41.859194, r.ParamFilterLatitude)
	}
	if r.ParamFilterLongitude != -87.646984 {
		s.t.Errorf("expected ParamLongitude to be %f, got: %f", -87.646984, r.ParamFilterLatitude)
	}
	if r.ParamFilterAppType != "fueling" {
		s.t.Errorf("expected ParamAppType to be %q, got: %q", "fueling", r.ParamFilterAppType)
	}

	appsResp := make(LocationBasedAppsWithRefs, 10)
	for i := 0; i < 10; i++ {
		appsResp[i] = &LocationBasedAppWithRefs{
			ID:                   strconv.Itoa(i),
			AndroidInstantAppURL: "https://foobar.com",
			Title:                "some app",
			AppType:              "some type",
		}
	}

	w.OK(appsResp)

	return nil
}
func (s *testService) GetApps(ctx context.Context, w GetAppsResponseWriter, r *GetAppsRequest) error {
	t := time.Date(2020, 5, 6, 12, 22, 54, 888456, time.UTC)
	if !r.ParamFilterSince.Equal(t) {
		s.t.Errorf("expected ParamFilterSince to be %q, got: %q", t, r.ParamFilterSince)
	}

	appsResp := make(LocationBasedApps, 10)
	for i := 0; i < 10; i++ {
		appsResp[i] = &LocationBasedApp{
			ID:                   strconv.Itoa(i),
			AndroidInstantAppURL: "https://foobar.com",
			Title:                "some app",
			AppType:              "some type",
		}
	}

	w.OK(appsResp)

	return nil
}
func (s *testService) CreateApp(context.Context, CreateAppResponseWriter, *CreateAppRequest) error {
	return nil
}
func (s *testService) DeleteApp(context.Context, DeleteAppResponseWriter, *DeleteAppRequest) error {
	return nil
}
func (s *testService) GetApp(context.Context, GetAppResponseWriter, *GetAppRequest) error { return nil }
func (s *testService) UpdateApp(context.Context, UpdateAppResponseWriter, *UpdateAppRequest) error {
	return nil
}
func (s *testService) GetAppPOIsRelationships(context.Context, GetAppPOIsRelationshipsResponseWriter, *GetAppPOIsRelationshipsRequest) error {
	return nil
}
func (s *testService) UpdateAppPOIsRelationships(context.Context, UpdateAppPOIsRelationshipsResponseWriter, *UpdateAppPOIsRelationshipsRequest) error {
	return nil
}
func (s *testService) GetDuplicatesKML(context.Context, GetDuplicatesKMLResponseWriter, *GetDuplicatesKMLRequest) error {
	return nil
}
func (s *testService) GetPoisDump(context.Context, GetPoisDumpResponseWriter, *GetPoisDumpRequest) error {
	return nil
}
func (s *testService) DeleteGasStationReferenceStatus(context.Context, DeleteGasStationReferenceStatusResponseWriter, *DeleteGasStationReferenceStatusRequest) error {
	return nil
}
func (s *testService) PutGasStationReferenceStatus(context.Context, PutGasStationReferenceStatusResponseWriter, *PutGasStationReferenceStatusRequest) error {
	return nil
}
func (s *testService) GetEvents(context.Context, GetEventsResponseWriter, *GetEventsRequest) error {
	return nil
}
func (s *testService) GetGasStations(context.Context, GetGasStationsResponseWriter, *GetGasStationsRequest) error {
	return nil
}
func (s *testService) GetGasStation(context.Context, GetGasStationResponseWriter, *GetGasStationRequest) error {
	return nil
}
func (s *testService) GetPriceHistory(context.Context, GetPriceHistoryResponseWriter, *GetPriceHistoryRequest) error {
	return nil
}
func (s *testService) GetGasStationFuelTypeNameMapping(context.Context, GetGasStationFuelTypeNameMappingResponseWriter, *GetGasStationFuelTypeNameMappingRequest) error {
	return nil
}
func (s *testService) GetMetadataFilters(context.Context, GetMetadataFiltersResponseWriter, *GetMetadataFiltersRequest) error {
	return nil
}
func (s *testService) GetPois(context.Context, GetPoisResponseWriter, *GetPoisRequest) error {
	return nil
}
func (s *testService) GetPoi(context.Context, GetPoiResponseWriter, *GetPoiRequest) error { return nil }
func (s *testService) ChangePoi(context.Context, ChangePoiResponseWriter, *ChangePoiRequest) error {
	return nil
}
func (s *testService) GetPolicies(context.Context, GetPoliciesResponseWriter, *GetPoliciesRequest) error {
	return nil
}
func (s *testService) CreatePolicy(ctx context.Context, w CreatePolicyResponseWriter, r *CreatePolicyRequest) error {
	if r.Content.Rules[0].Priorities[0].SourceID != "f106ac99-213c-4cf7-8c1b-1e841516026b" {
		return fmt.Errorf("Expected first rule first property source id to equal %q", "f106ac99-213c-4cf7-8c1b-1e841516026b")
	}

	return nil
}
func (s *testService) GetPolicy(context.Context, GetPolicyResponseWriter, *GetPolicyRequest) error {
	return nil
}
func (s *testService) GetRegionalPrices(context.Context, GetRegionalPricesResponseWriter, *GetRegionalPricesRequest) error {
	return nil
}
func (s *testService) GetSources(context.Context, GetSourcesResponseWriter, *GetSourcesRequest) error {
	return nil
}
func (s *testService) CreateSource(context.Context, CreateSourceResponseWriter, *CreateSourceRequest) error {
	return nil
}
func (s *testService) DeleteSource(context.Context, DeleteSourceResponseWriter, *DeleteSourceRequest) error {
	return nil
}
func (s *testService) GetSource(context.Context, GetSourceResponseWriter, *GetSourceRequest) error {
	return nil
}
func (s *testService) UpdateSource(context.Context, UpdateSourceResponseWriter, *UpdateSourceRequest) error {
	return nil
}
func (s *testService) GetSubscriptions(context.Context, GetSubscriptionsResponseWriter, *GetSubscriptionsRequest) error {
	return nil
}

func (s *testService) DeleteSubscription(context.Context, DeleteSubscriptionResponseWriter, *DeleteSubscriptionRequest) error {
	return nil
}

func (s *testService) StoreSubscription(context.Context, StoreSubscriptionResponseWriter, *StoreSubscriptionRequest) error {
	return nil
}

func (s *testService) GetTiles(context.Context, GetTilesResponseWriter, *GetTilesRequest) error {
	return nil
}

type testAuthBackend struct {
}

func (s testAuthBackend) AuthorizeDeviceID(r *http.Request, w http.ResponseWriter) (context.Context, bool) {
	return r.Context(), true
}

func (s testAuthBackend) InitDeviceID(cfgDeviceID *apikey.Config) {}

func (s testAuthBackend) AuthorizeOAuth2(r *http.Request, w http.ResponseWriter, scope string) (context.Context, bool) {
	return r.Context(), true
}

func (s testAuthBackend) InitOAuth2(cfgOAuth2 *oauth2.Config) {}

func (s testAuthBackend) AuthorizeOIDC(r *http.Request, w http.ResponseWriter, scope string) (context.Context, bool) {
	return r.Context(), true
}

func (s testAuthBackend) InitOIDC(cfgOIDC *oidc.Config) {}

func TestHandler(t *testing.T) {
	r := Router(&testService{t}, &testAuthBackend{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/poi/beta/apps/query?"+
		"filter[latitude]=41.859194&filter[longitude]=-87.646984&filter[appType]=fueling", nil)
	req.Header.Set("Accept", runtime.JSONAPIContentType)
	req.Header.Set("Content-Type", runtime.JSONAPIContentType)

	r.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected OK got: %d", resp.StatusCode)
		t.Error(rec.Body.String())
		return
	}

	var data struct {
		Data []map[string]interface{} `json:"data"`
	}
	err := json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		t.Fatal(err)
		return
	}
	if len(data.Data) != 10 {
		t.Error("Expected 10 apps")
		return
	}
	if data.Data[0]["type"] != "locationBasedAppWithRefs" {
		t.Error("Expected type locationBasedAppWithRefs")
		return
	}
	attributes, ok := data.Data[0]["attributes"].(map[string]interface{})
	if !ok {
		t.Error("Expected attributes do be present")
		return
	}
	if attributes["androidInstantAppUrl"] != "https://foobar.com" {
		t.Error(`Expected androidInstantAppUrl to be "https://foobar.com"`)
	}
	if attributes["title"] != "some app" {
		t.Error(`Expected androidInstantAppUrl to be "some app"`)
	}
	if attributes["appType"] != "some type" {
		t.Error(`Expected androidInstantAppUrl to be "some type"`)
	}
}

func TestHandlerWithTimeInQuery(t *testing.T) {
	r := Router(&testService{t}, &testAuthBackend{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/poi/beta/apps?filter[since]=2020-05-06T12%3A22%3A54%2E000888456", nil)
	req.Header.Set("Accept", runtime.JSONAPIContentType)
	req.Header.Set("Content-Type", runtime.JSONAPIContentType)

	r.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected OK got: %d", resp.StatusCode)
		t.Error(rec.Body.String())
		return
	}

	var data struct {
		Data []map[string]interface{} `json:"data"`
	}
	err := json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		t.Fatal(err)
		return
	}
	if len(data.Data) != 10 {
		t.Error("Expected 10 apps")
		return
	}
	if data.Data[0]["type"] != "locationBasedApp" {
		t.Error("Expected type locationBasedApp")
		return
	}
	attributes, ok := data.Data[0]["attributes"].(map[string]interface{})
	if !ok {
		t.Error("Expected attributes do be present")
		return
	}
	if attributes["androidInstantAppUrl"] != "https://foobar.com" {
		t.Error(`Expected androidInstantAppUrl to be "https://foobar.com"`)
	}
	if attributes["title"] != "some app" {
		t.Error(`Expected androidInstantAppUrl to be "some app"`)
	}
	if attributes["appType"] != "some type" {
		t.Error(`Expected androidInstantAppUrl to be "some type"`)
	}
}

func TestCreatePolicyHandler(t *testing.T) {
	r := Router(&testService{t}, &testAuthBackend{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/poi/beta/policies", strings.NewReader(`{
		"data": {
			"id": "f106ac99-213c-4cf7-8c1b-1e841516026b",
			"type": "policies",
			"attributes": {
				"poiType": "SpeedCamera",
				"userId": "f106ac99-213c-4cf7-8c1b-1e841516026b",
				"countryId": "DE",
				"createdAt": "2018-01-01T00:00:00Z",
				"rules": [
					{
						"field": "name",
						"priorities": [
							{
								"sourceId": "f106ac99-213c-4cf7-8c1b-1e841516026b",
								"timeToLive": 0
							}
						]
					}
				]
			}
		}
	}`))
	req.Header.Set("Accept", runtime.JSONAPIContentType)
	req.Header.Set("Content-Type", runtime.JSONAPIContentType)

	r.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected OK got: %d", resp.StatusCode)
		t.Error(rec.Body.String())
		return
	}
}
