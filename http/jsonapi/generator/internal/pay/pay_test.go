package pay

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/pace/bricks/http/jsonapi"
	"github.com/pace/bricks/http/jsonapi/runtime"
	"github.com/pace/bricks/http/oauth2"
	oidc "github.com/pace/bricks/http/oidc"
	"github.com/pace/bricks/http/security/apikey"
	"github.com/pace/bricks/maintenance/log"
)

type testService struct {
	t *testing.T
}

func (s *testService) GetPaymentMethods(context.Context, GetPaymentMethodsResponseWriter, *GetPaymentMethodsRequest) error {
	panic("Some error!")
}

func (s *testService) CreatePaymentMethodSEPA(ctx context.Context, w CreatePaymentMethodSEPAResponseWriter, r *CreatePaymentMethodSEPARequest) error {
	if str := "Jon"; r.Content.FirstName != str {
		s.t.Errorf("expected FirstName to be %q, got %q", str, r.Content.FirstName)
	}
	if str := "Haid-und-Neu-Str."; r.Content.Address.Street != str {
		s.t.Errorf("expected Address.Street to be %q, got %q", str, r.Content.Address.Street)
	}

	w.Created(&CreatePaymentMethodSEPACreated{
		ID:                   "1",
		IdentificationString: "d7101f72-a672-453c-9d36-d5809ef0ded6",
		Kind:                 "sepa",
	})

	return nil
}

func (s *testService) DeletePaymentMethod(context.Context, DeletePaymentMethodResponseWriter, *DeletePaymentMethodRequest) error {
	return nil
}

func (s *testService) AuthorizePaymentMethod(context.Context, AuthorizePaymentMethodResponseWriter, *AuthorizePaymentMethodRequest) error {
	return nil
}

func (s *testService) DeletePaymentToken(context.Context, DeletePaymentTokenResponseWriter, *DeletePaymentTokenRequest) error {
	return nil
}

func (s *testService) GetPaymentMethodsIncludingCreditCheck(context.Context, GetPaymentMethodsIncludingCreditCheckResponseWriter, *GetPaymentMethodsIncludingCreditCheckRequest) error {
	return nil
}

func (s *testService) GetPaymentMethodsIncludingPaymentToken(context.Context, GetPaymentMethodsIncludingPaymentTokenResponseWriter, *GetPaymentMethodsIncludingPaymentTokenRequest) error {
	return fmt.Errorf("Some other error")
}

func (s *testService) ProcessPayment(ctx context.Context, w ProcessPaymentResponseWriter, r *ProcessPaymentRequest) error {
	if r.ParamPathDecimal.String() != "1337.42" {
		s.t.Errorf(`expected pathDecimal "1337.42", got %q`, r.ParamPathDecimal)
	}

	if r.ParamQueryDecimal.String() != "123.456" {
		s.t.Errorf(`expected queryDecimal "123.456", got %q`, r.ParamPathDecimal)
	}

	if r.Content.PriceIncludingVAT.String() != "69.34" {
		s.t.Errorf(`expected priceIncludingVAT "69.34", got %q`, r.Content.PriceIncludingVAT)
	}
	amount := decimal.RequireFromString("11.07")
	rate := decimal.RequireFromString("19.0")
	priceWithVat := decimal.RequireFromString("69.34")
	priceWithoutVat := decimal.RequireFromString("58.27")
	w.Created(&ProcessPaymentCreated{
		ID: "42",
		VAT: ProcessPaymentCreatedVAT{
			Amount: &amount,
			Rate:   &rate,
		},
		Currency: "EUR",
		Fueling: ProcessPaymentCreatedFueling{
			AppID:   "c30bce97-b732-4390-af38-1ac6b017aa4c",
			PumpID:  "460ffaad-a3c1-4199-b69e-63949ccda82f",
			Vin:     "1B3EL46R36N102271",
			Mileage: 66435,
		},
		PaymentToken:      "f106ac99-213c-4cf7-8c1b-1e841516026b",
		PriceIncludingVAT: &priceWithVat,
		PriceWithoutVAT:   &priceWithoutVat,
	})

	return nil
}

type testAuthBackend struct{}

func (s testAuthBackend) CanAuthorizeOAuth2(r *http.Request) bool {
	return true
}

func (s testAuthBackend) CanAuthorizeOpenID(r *http.Request) bool {
	return true
}

func (s testAuthBackend) CanAuthorizeProfileKey(r *http.Request) bool {
	return true
}

func (s testAuthBackend) AuthorizeOAuth2(r *http.Request, w http.ResponseWriter, scope string) (context.Context, bool) {
	return r.Context(), true
}

func (s testAuthBackend) AuthorizeOpenID(r *http.Request, w http.ResponseWriter, scope string) (context.Context, bool) {
	return r.Context(), true
}

func (s testAuthBackend) AuthorizeProfileKey(r *http.Request, w http.ResponseWriter) (context.Context, bool) {
	return r.Context(), true
}

func (s testAuthBackend) InitOAuth2(cfgOAuth2 *oauth2.Config) {
}

func (s testAuthBackend) InitOpenID(cfgOpenID *oidc.Config) {
}

func (s testAuthBackend) InitProfileKey(cfgProfileKey *apikey.Config) {
}

func TestHandler(t *testing.T) {
	r := Router(&testService{t}, &testAuthBackend{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/pay/beta/payment-methods/sepa-direct-debit", strings.NewReader(`{
		"data": {
			"id": "2a1319c3-c136-495d-b59a-47b3246d08af",
			"type": "paymentMethod",
			"attributes": {
				"kind": "sepa",
				"iban": "DE89370400440532013000",
				"firstName": "Jon",
				"lastName": "Smith",
				"address": {
					"street": "Haid-und-Neu-Str.",
					"houseNo": "18",
					"postalCode": "76131",
					"city": "Karlsruhe",
					"countryCode": "DE"
				}
			}
		}
	}`))
	req.Header.Set("Accept", runtime.JSONAPIContentType)
	req.Header.Set("Content-Type", runtime.JSONAPIContentType)

	r.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		t.Errorf("expected OK got: %d", resp.StatusCode)
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		t.Error(string(b[:]))
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(b[:]), "d7101f72-a672-453c-9d36-d5809ef0ded6") {
		t.Error("expected response to contain the generated payment method id")
	}
}

func TestHandlerDecimal(t *testing.T) {
	r := Router(&testService{t}, &testAuthBackend{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/pay/beta/transaction/1337.42?queryDecimal=123.456", strings.NewReader(`{
		"data": {
			"id": "5d3607f4-7855-4bfc-b926-1e662c225f06",
			"type": "transaction",
			"attributes": {
				"paymentToken": "f106ac99-213c-4cf7-8c1b-1e841516026b",
				"fueling": {
	 				"appId": "c30bce97-b732-4390-af38-1ac6b017aa4c",
	 				"pumpId": "460ffaad-a3c1-4199-b69e-63949ccda82f",
	 				"vin": "1B3EL46R36N102271",
	 				"mileage": 66435
				},
				"currency": "EUR",
				"priceIncludingVAT": 69.34
			}
		}
	}`))
	req.Header.Set("Accept", runtime.JSONAPIContentType)
	req.Header.Set("Content-Type", runtime.JSONAPIContentType)

	r.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		t.Errorf("expected OK got: %d", resp.StatusCode)
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		t.Error(string(b[:]))
	}

	var pc ProcessPaymentCreated
	if err := jsonapi.UnmarshalPayload(resp.Body, &pc); err != nil {
		t.Fatal(err)
	}

	assertDecimal(t, *pc.VAT.Amount, decimal.RequireFromString("11.07"))
	assertDecimal(t, *pc.VAT.Rate, decimal.RequireFromString("19.0"))
	assertDecimal(t, *pc.PriceIncludingVAT, decimal.RequireFromString("69.34"))
	assertDecimal(t, *pc.PriceWithoutVAT, decimal.RequireFromString("58.27"))
}

func assertDecimal(t *testing.T, got, want decimal.Decimal) {
	if !got.Equals(want) {
		t.Errorf(`expected decimal.Decimal %q, got %q`, want, got)
	}
}

func TestHandlerPanic(t *testing.T) {
	r := Router(&testService{t}, &testAuthBackend{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/pay/beta/payment-methods?include=paymentToken", nil)
	req.Header.Set("Accept", runtime.JSONAPIContentType)
	req.Header.Set("Content-Type", runtime.JSONAPIContentType)

	log.Handler()(r).ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500 got: %d", resp.StatusCode)
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		t.Error(string(b[:]))
	}
}

func TestHandlerError(t *testing.T) {
	r := Router(&testService{t}, &testAuthBackend{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/pay/beta/payment-methods", nil)
	req.Header.Set("Accept", runtime.JSONAPIContentType)
	req.Header.Set("Content-Type", runtime.JSONAPIContentType)

	log.Handler()(r).ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500 got: %d", resp.StatusCode)
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		t.Error(string(b[:]))
	}
}
