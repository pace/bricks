package fueling

import (
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pace/bricks/http/jsonapi/runtime"
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
	b, _ := io.ReadAll(resp.Body)

	require.Equalf(t, 422, resp.StatusCode, "expected 422 got: %s", string(b))
	assert.Contains(t, string(b), `can't parse content: got value \"47.8\" expected type float32: Invalid type provided`)
}
