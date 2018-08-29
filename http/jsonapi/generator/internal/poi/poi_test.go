package poi

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
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

	return fmt.Errorf("test")
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

	if resp.StatusCode != 500 {
		t.Errorf("expected OK got: %d", resp.StatusCode)
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		t.Error(string(b[:]))
	}
}
