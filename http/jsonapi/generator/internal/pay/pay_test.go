package pay

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pace/bricks/http/jsonapi/runtime"
	"github.com/pace/bricks/http/oauth2"
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

func (s *testService) ProcessPayment(context.Context, ProcessPaymentResponseWriter, *ProcessPaymentRequest) error {
	return nil
}

type testAuthBackend struct {
}

func (s testAuthBackend) AuthorizeOAuth2(r *http.Request, w http.ResponseWriter, scope string) (context.Context, bool) {
	return r.Context(), true
}

func (s testAuthBackend) AuthorizeProfileKey(r *http.Request, w http.ResponseWriter) (context.Context, bool) {
	return r.Context(), true
}

func (s testAuthBackend) Init(cfgOAuth2 *oauth2.Config, cfgProfileKey *apikey.Config) {
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

	if !strings.Contains(string(b[:]), "d7101f72-a672-453c-9d36-d5809ef0ded6") {
		t.Error("expected response to contain the generated payment method id")
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
		b, err := ioutil.ReadAll(resp.Body)
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
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		t.Error(string(b[:]))
	}
}
