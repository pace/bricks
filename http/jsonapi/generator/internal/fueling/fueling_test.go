package fueling

import (
	"context"
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"

	runtime "github.com/pace/bricks/http/jsonapi/runtime"
)

type testService struct {
	t *testing.T
}

func (t *testService) ProcessPayment(context.Context, ProcessPaymentResponseWriter, *ProcessPaymentRequest) error {
	return nil
}

func (t *testService) ApproachingAtTheForecourt(context.Context, ApproachingAtTheForecourtResponseWriter, *ApproachingAtTheForecourtRequest) error {
	return nil
}

func (t *testService) GetPump(context.Context, GetPumpResponseWriter, *GetPumpRequest) error {
	return nil
}

func (t *testService) WaitOnPumpStatusChange(context.Context, WaitOnPumpStatusChangeResponseWriter, *WaitOnPumpStatusChangeRequest) error {
	return nil
}

func TestErrorReporting(t *testing.T) {
	r := Router(&testService{t})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/fueling/beta/gas-stations/d7101f72-a672-453c-9d36-d5809ef0ded6/approaching", strings.NewReader(`{
		"data": {
		  "type": "approaching",
		  "id": "c3f037ea-492e-4033-9b4b-4efc7beca16c",
		  "attributes": {
			"expectedAmount": "47.8",
			"carFuelType": "ron95_e10"
		  }
		}
	  }`))
	req.Header.Set("Accept", runtime.JSONAPIContentType)
	req.Header.Set("Content-Type", runtime.JSONAPIContentType)

	r.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != 422 {
		t.Errorf("expected 422 got: %d", resp.StatusCode)
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		t.Error(string(b[:]))
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(b[:]), "can't parse content: Got value \"47.8\" expected type float32: Invalid type provided") {
		t.Errorf("expected response to contain a better error message, got: %s", string(b[:]))
	}
}
