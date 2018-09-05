package pay

import (
	"context"
	"errors"
	"fmt"
	mux "github.com/gorilla/mux"
	runtime "lab.jamit.de/pace/go-microservice/http/jsonapi/runtime"
	jsonapimetrics "lab.jamit.de/pace/go-microservice/maintenance/metrics/jsonapi"
	"net/http"
	"runtime/debug"
)

// AllPaymentMethodsItem ...
type AllPaymentMethodsItem struct {
	ID                   string `jsonapi:"primary,paymentMethod,omitempty" valid:"uuid,optional"`                                      // Payment method ID
	IdentificationString string `json:"identificationString,omitempty" jsonapi:"attr,identificationString,omitempty" valid:"optional"` // Example: "DE89 **** 3000"
	Kind                 string `json:"kind,omitempty" jsonapi:"attr,kind,omitempty" valid:"optional"`                                 // Example: "sepa"
}

// AllPaymentMethods ...
type AllPaymentMethods []*AllPaymentMethodsItem

// PaymentMethodSEPAAddress ...
type PaymentMethodSEPAAddress struct {
	City        string `json:"city,omitempty" jsonapi:"city,omitempty" valid:"required"`               // Example: "Karlsruhe"
	CountryCode string `json:"countryCode,omitempty" jsonapi:"countryCode,omitempty" valid:"required"` // Country code in as specified in ISO 3166-1.
	HouseNo     string `json:"houseNo,omitempty" jsonapi:"houseNo,omitempty" valid:"required"`         // Example: "18"
	PostalCode  string `json:"postalCode,omitempty" jsonapi:"postalCode,omitempty" valid:"required"`   // Example: "76131"
	Street      string `json:"street,omitempty" jsonapi:"street,omitempty" valid:"required"`           // Example: "Haid-und-Neu-Str."
}

// PaymentMethodSEPA ...
type PaymentMethodSEPA struct {
	ID        string                    `jsonapi:"primary,paymentMethod,omitempty" valid:"uuid,optional"` // The ID of this payment method.
	Address   *PaymentMethodSEPAAddress `json:"address,omitempty" jsonapi:"attr,address,omitempty" valid:"required"`
	FirstName string                    `json:"firstName,omitempty" jsonapi:"attr,firstName,omitempty" valid:"required"` // Example: "Jon"
	Iban      string                    `json:"iban,omitempty" jsonapi:"attr,iban,omitempty" valid:"required"`           // Example: "DE89370400440532013000"
	Kind      string                    `json:"kind,omitempty" jsonapi:"attr,kind,omitempty" valid:"required"`
	LastName  string                    `json:"lastName,omitempty" jsonapi:"attr,lastName,omitempty" valid:"required"` // Example: "Smith"
}

// PaymentMethodsWithPaymentTokensItem ...
type PaymentMethodsWithPaymentTokensItem struct {
	ID                   string `jsonapi:"primary,paymentMethod,omitempty" valid:"uuid,optional"`                                      // Payment method ID
	IdentificationString string `json:"identificationString,omitempty" jsonapi:"attr,identificationString,omitempty" valid:"optional"` // Example: "DE89 **** 3000"
	Kind                 string `json:"kind,omitempty" jsonapi:"attr,kind,omitempty" valid:"optional"`                                 // Example: "sepa"
}

// PaymentMethodsWithPaymentTokens ...
type PaymentMethodsWithPaymentTokens []*PaymentMethodsWithPaymentTokensItem

/*
GetPaymentMethodsHandler handles request/response marshaling and validation for
 Get /beta/payment-methods
*/
func GetPaymentMethodsHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Panic %s: %v\n", "GetPaymentMethodsHandler", r)
				debug.PrintStack()
				runtime.WriteError(w, http.StatusInternalServerError, errors.New("Error"))
			}
		}()
		writer := getPaymentMethodsResponseWriter{
			ResponseWriter: jsonapimetrics.NewMetric("pay", "/beta/payment-methods", w, r),
		}
		request := GetPaymentMethodsRequest{
			Request: r,
		}
		err := service.GetPaymentMethods(r.Context(), &writer, &request)
		if err != nil {
			runtime.WriteError(w, http.StatusInternalServerError, err)
		}
	})
}

/*
CreatePaymentMethodSEPAHandler handles request/response marshaling and validation for
 Post /beta/payment-methods/sepa-direct-debit
*/
func CreatePaymentMethodSEPAHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Panic %s: %v\n", "CreatePaymentMethodSEPAHandler", r)
				debug.PrintStack()
				runtime.WriteError(w, http.StatusInternalServerError, errors.New("Error"))
			}
		}()
		writer := createPaymentMethodSEPAResponseWriter{
			ResponseWriter: jsonapimetrics.NewMetric("pay", "/beta/payment-methods/sepa-direct-debit", w, r),
		}
		request := CreatePaymentMethodSEPARequest{
			Request: r,
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}
		if runtime.Unmarshal(w, r, &request.Content) {
			err := service.CreatePaymentMethodSEPA(r.Context(), &writer, &request)
			if err != nil {
				runtime.WriteError(w, http.StatusInternalServerError, err)
			}
		}
	})
}

/*
DeletePaymentMethodHandler handles request/response marshaling and validation for
 Delete /beta/payment-methods/{paymentMethodId}
*/
func DeletePaymentMethodHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Panic %s: %v\n", "DeletePaymentMethodHandler", r)
				debug.PrintStack()
				runtime.WriteError(w, http.StatusInternalServerError, errors.New("Error"))
			}
		}()
		writer := deletePaymentMethodResponseWriter{
			ResponseWriter: jsonapimetrics.NewMetric("pay", "/beta/payment-methods/{paymentMethodId}", w, r),
		}
		request := DeletePaymentMethodRequest{
			Request: r,
		}
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamPaymentMethodID,
			Location: runtime.ScanInPath,
			Input:    vars["paymentMethodId"],
			Name:     "paymentMethodId",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}
		err := service.DeletePaymentMethod(r.Context(), &writer, &request)
		if err != nil {
			runtime.WriteError(w, http.StatusInternalServerError, err)
		}
	})
}

/*
AuthorizePaymentMethodHandler handles request/response marshaling and validation for
 Post /beta/payment-methods/{paymentMethodId}/authorize
*/
func AuthorizePaymentMethodHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Panic %s: %v\n", "AuthorizePaymentMethodHandler", r)
				debug.PrintStack()
				runtime.WriteError(w, http.StatusInternalServerError, errors.New("Error"))
			}
		}()
		writer := authorizePaymentMethodResponseWriter{
			ResponseWriter: jsonapimetrics.NewMetric("pay", "/beta/payment-methods/{paymentMethodId}/authorize", w, r),
		}
		request := AuthorizePaymentMethodRequest{
			Request: r,
		}
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamPaymentMethodID,
			Location: runtime.ScanInPath,
			Input:    vars["paymentMethodId"],
			Name:     "paymentMethodId",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}
		if runtime.Unmarshal(w, r, &request.Content) {
			err := service.AuthorizePaymentMethod(r.Context(), &writer, &request)
			if err != nil {
				runtime.WriteError(w, http.StatusInternalServerError, err)
			}
		}
	})
}

/*
DeletePaymentTokenHandler handles request/response marshaling and validation for
 Delete /beta/payment-methods/{paymentMethodId}/paymentTokens/{paymentTokenId}
*/
func DeletePaymentTokenHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Panic %s: %v\n", "DeletePaymentTokenHandler", r)
				debug.PrintStack()
				runtime.WriteError(w, http.StatusInternalServerError, errors.New("Error"))
			}
		}()
		writer := deletePaymentTokenResponseWriter{
			ResponseWriter: jsonapimetrics.NewMetric("pay", "/beta/payment-methods/{paymentMethodId}/paymentTokens/{paymentTokenId}", w, r),
		}
		request := DeletePaymentTokenRequest{
			Request: r,
		}
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamPaymentTokenID,
			Location: runtime.ScanInPath,
			Input:    vars["paymentTokenId"],
			Name:     "paymentTokenId",
		}, &runtime.ScanParameter{
			Data:     &request.ParamPaymentMethodID,
			Location: runtime.ScanInPath,
			Input:    vars["paymentMethodId"],
			Name:     "paymentMethodId",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}
		err := service.DeletePaymentToken(r.Context(), &writer, &request)
		if err != nil {
			runtime.WriteError(w, http.StatusInternalServerError, err)
		}
	})
}

/*
GetPaymentMethodsIncludingCreditCheckHandler handles request/response marshaling and validation for
 Get /beta/payment-methods?include=creditCheck
*/
func GetPaymentMethodsIncludingCreditCheckHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Panic %s: %v\n", "GetPaymentMethodsIncludingCreditCheckHandler", r)
				debug.PrintStack()
				runtime.WriteError(w, http.StatusInternalServerError, errors.New("Error"))
			}
		}()
		writer := getPaymentMethodsIncludingCreditCheckResponseWriter{
			ResponseWriter: jsonapimetrics.NewMetric("pay", "/beta/payment-methods?include=creditCheck", w, r),
		}
		request := GetPaymentMethodsIncludingCreditCheckRequest{
			Request: r,
		}
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamInclude,
			Location: runtime.ScanInQuery,
			Input:    vars["include"],
			Name:     "include",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}
		err := service.GetPaymentMethodsIncludingCreditCheck(r.Context(), &writer, &request)
		if err != nil {
			runtime.WriteError(w, http.StatusInternalServerError, err)
		}
	})
}

/*
GetPaymentMethodsIncludingPaymentTokenHandler handles request/response marshaling and validation for
 Get /beta/payment-methods?include=paymentTokens
*/
func GetPaymentMethodsIncludingPaymentTokenHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Panic %s: %v\n", "GetPaymentMethodsIncludingPaymentTokenHandler", r)
				debug.PrintStack()
				runtime.WriteError(w, http.StatusInternalServerError, errors.New("Error"))
			}
		}()
		writer := getPaymentMethodsIncludingPaymentTokenResponseWriter{
			ResponseWriter: jsonapimetrics.NewMetric("pay", "/beta/payment-methods?include=paymentTokens", w, r),
		}
		request := GetPaymentMethodsIncludingPaymentTokenRequest{
			Request: r,
		}
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamInclude,
			Location: runtime.ScanInQuery,
			Input:    vars["include"],
			Name:     "include",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}
		err := service.GetPaymentMethodsIncludingPaymentToken(r.Context(), &writer, &request)
		if err != nil {
			runtime.WriteError(w, http.StatusInternalServerError, err)
		}
	})
}

/*
GetPaymentMethodsResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetPaymentMethodsResponseWriter interface {
	http.ResponseWriter
	AllThePaymentMethodsForUser(AllPaymentMethods)
}
type getPaymentMethodsResponseWriter struct {
	http.ResponseWriter
}

// AllThePaymentMethodsForUser responds with jsonapi marshaled data (HTTP code 200)
func (w *getPaymentMethodsResponseWriter) AllThePaymentMethodsForUser(data AllPaymentMethods) {
	runtime.Marshal(w, data, 200)
}

/*
GetPaymentMethodsResponseWriter is a standard http.Request extended with the
un-marshaled content object
*/
type GetPaymentMethodsRequest struct {
	Request *http.Request `valid:"-"`
}

// CreatePaymentMethodSEPACreated ...
type CreatePaymentMethodSEPACreated struct {
	ID                   string `jsonapi:"primary,paymentMethod,omitempty" valid:"uuid,optional"`                                      // Payment method ID
	IdentificationString string `json:"identificationString,omitempty" jsonapi:"attr,identificationString,omitempty" valid:"optional"` // Example: "DE89 **** 3000"
	Kind                 string `json:"kind,omitempty" jsonapi:"attr,kind,omitempty" valid:"optional"`
}

/*
CreatePaymentMethodSEPAResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type CreatePaymentMethodSEPAResponseWriter interface {
	http.ResponseWriter
	Created(*CreatePaymentMethodSEPACreated)
	BadRequest(error)
}
type createPaymentMethodSEPAResponseWriter struct {
	http.ResponseWriter
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *createPaymentMethodSEPAResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// Created responds with jsonapi marshaled data (HTTP code 201)
func (w *createPaymentMethodSEPAResponseWriter) Created(data *CreatePaymentMethodSEPACreated) {
	runtime.Marshal(w, data, 201)
}

// CreatePaymentMethodSEPARequest ...
type CreatePaymentMethodSEPARequest struct {
	Request *http.Request     `valid:"-"`
	Content PaymentMethodSEPA `valid:"-"`
}

/*
DeletePaymentMethodResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type DeletePaymentMethodResponseWriter interface {
	http.ResponseWriter
	ThePaymentMethodWasDeletedSuccessfully()
	NotFound(error)
}
type deletePaymentMethodResponseWriter struct {
	http.ResponseWriter
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *deletePaymentMethodResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// ThePaymentMethodWasDeletedSuccessfully responds with empty response (HTTP code 204)
func (w *deletePaymentMethodResponseWriter) ThePaymentMethodWasDeletedSuccessfully() {
	w.WriteHeader(204)
}

/*
DeletePaymentMethodResponseWriter is a standard http.Request extended with the
un-marshaled content object
*/
type DeletePaymentMethodRequest struct {
	Request              *http.Request `valid:"-"`
	ParamPaymentMethodID string        `valid:"required,uuid"`
}

// AuthorizePaymentMethodOK ...
type AuthorizePaymentMethodOK struct {
	ID       string  `jsonapi:"primary,paymentToken,omitempty" valid:"uuid,optional"`               // paymentToken ID (NOT the token value)
	Amount   float64 `json:"amount,omitempty" jsonapi:"attr,amount,omitempty" valid:"optional"`     // Example: "65.49"
	Currency string  `json:"currency,omitempty" jsonapi:"attr,currency,omitempty" valid:"optional"` // Currency as specified in ISO-4217.
	Value    string  `json:"value,omitempty" jsonapi:"attr,value,omitempty" valid:"optional"`       // The actual token value. Note that the format is subject to change. Treat transparently.
}

/*
AuthorizePaymentMethodResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type AuthorizePaymentMethodResponseWriter interface {
	http.ResponseWriter
	OK(*AuthorizePaymentMethodOK)
	AmountCannotBeAuthorized(error)
	PaymentMethodIsUnknown(error)
	BadGateway(error)
}
type authorizePaymentMethodResponseWriter struct {
	http.ResponseWriter
}

// BadGateway responds with jsonapi error (HTTP code 502)
func (w *authorizePaymentMethodResponseWriter) BadGateway(err error) {
	runtime.WriteError(w, 502, err)
}

// PaymentMethodIsUnknown responds with jsonapi error (HTTP code 404)
func (w *authorizePaymentMethodResponseWriter) PaymentMethodIsUnknown(err error) {
	runtime.WriteError(w, 404, err)
}

// AmountCannotBeAuthorized responds with jsonapi error (HTTP code 403)
func (w *authorizePaymentMethodResponseWriter) AmountCannotBeAuthorized(err error) {
	runtime.WriteError(w, 403, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *authorizePaymentMethodResponseWriter) OK(data *AuthorizePaymentMethodOK) {
	runtime.Marshal(w, data, 200)
}

// AuthorizePaymentMethodContent ...
type AuthorizePaymentMethodContent struct {
	ID       string  `jsonapi:"primary,paymentToken,omitempty" valid:"uuid,optional"`               // ID of the new paymentToken.
	Amount   float64 `json:"amount,omitempty" jsonapi:"attr,amount,omitempty" valid:"required"`     // Example: "65.49"
	Currency string  `json:"currency,omitempty" jsonapi:"attr,currency,omitempty" valid:"required"` // Currency as specified in ISO-4217.
}

// AuthorizePaymentMethodRequest ...
type AuthorizePaymentMethodRequest struct {
	Request              *http.Request                  `valid:"-"`
	Content              *AuthorizePaymentMethodContent `valid:"-"`
	ParamPaymentMethodID string                         `valid:"required,uuid"`
}

/*
DeletePaymentTokenResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type DeletePaymentTokenResponseWriter interface {
	http.ResponseWriter
	ThePaymentTokenWasRemovedSuccessfully()
	NotFound(error)
}
type deletePaymentTokenResponseWriter struct {
	http.ResponseWriter
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *deletePaymentTokenResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// ThePaymentTokenWasRemovedSuccessfully responds with empty response (HTTP code 204)
func (w *deletePaymentTokenResponseWriter) ThePaymentTokenWasRemovedSuccessfully() {
	w.WriteHeader(204)
}

/*
DeletePaymentTokenResponseWriter is a standard http.Request extended with the
un-marshaled content object
*/
type DeletePaymentTokenRequest struct {
	Request              *http.Request `valid:"-"`
	ParamPaymentTokenID  string        `valid:"required"`
	ParamPaymentMethodID string        `valid:"required,uuid"`
}

/*
GetPaymentMethodsIncludingCreditCheckResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetPaymentMethodsIncludingCreditCheckResponseWriter interface {
	http.ResponseWriter
	AllThePaymentMethodsThatCouldBeUsed(AllPaymentMethods)
}
type getPaymentMethodsIncludingCreditCheckResponseWriter struct {
	http.ResponseWriter
}

// AllThePaymentMethodsThatCouldBeUsed responds with jsonapi marshaled data (HTTP code 200)
func (w *getPaymentMethodsIncludingCreditCheckResponseWriter) AllThePaymentMethodsThatCouldBeUsed(data AllPaymentMethods) {
	runtime.Marshal(w, data, 200)
}

/*
GetPaymentMethodsIncludingCreditCheckResponseWriter is a standard http.Request extended with the
un-marshaled content object
*/
type GetPaymentMethodsIncludingCreditCheckRequest struct {
	Request      *http.Request `valid:"-"`
	ParamInclude string        `valid:"required,in(creditCheck)"`
}

/*
GetPaymentMethodsIncludingPaymentTokenResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetPaymentMethodsIncludingPaymentTokenResponseWriter interface {
	http.ResponseWriter
	AllThePaymentMethodsWithPreAuthorisedAmounts(PaymentMethodsWithPaymentTokens)
}
type getPaymentMethodsIncludingPaymentTokenResponseWriter struct {
	http.ResponseWriter
}

// AllThePaymentMethodsWithPreAuthorisedAmounts responds with jsonapi marshaled data (HTTP code 200)
func (w *getPaymentMethodsIncludingPaymentTokenResponseWriter) AllThePaymentMethodsWithPreAuthorisedAmounts(data PaymentMethodsWithPaymentTokens) {
	runtime.Marshal(w, data, 200)
}

/*
GetPaymentMethodsIncludingPaymentTokenResponseWriter is a standard http.Request extended with the
un-marshaled content object
*/
type GetPaymentMethodsIncludingPaymentTokenRequest struct {
	Request      *http.Request `valid:"-"`
	ParamInclude string        `valid:"required,in(paymentToken)"`
}
type Service interface {
	// GetPaymentMethods Get all payment methods for user
	GetPaymentMethods(context.Context, GetPaymentMethodsResponseWriter, *GetPaymentMethodsRequest) error
	/*
	   CreatePaymentMethodSEPA Register SEPA direct debit as a payment method

	   By registering you allow the user to use SEPA direct debit as a payment method.
	   The payment method ID is optional when posting data.
	*/
	CreatePaymentMethodSEPA(context.Context, CreatePaymentMethodSEPAResponseWriter, *CreatePaymentMethodSEPARequest) error
	// DeletePaymentMethod Delete a payment method
	DeletePaymentMethod(context.Context, DeletePaymentMethodResponseWriter, *DeletePaymentMethodRequest) error
	/*
	   AuthorizePaymentMethod Authorize a payment using the payment method whose ID is paymentMethodId

	   When successful, returns a paymentToken value.
	*/
	AuthorizePaymentMethod(context.Context, AuthorizePaymentMethodResponseWriter, *AuthorizePaymentMethodRequest) error
	// DeletePaymentToken Delete the paymentToken record.
	DeletePaymentToken(context.Context, DeletePaymentTokenResponseWriter, *DeletePaymentTokenRequest) error
	/*
	   GetPaymentMethodsIncludingCreditCheck Get all ready-to-use payment methods for user

	   This request will return a list of supported payment methods for the current user that they can, in theory, use. That is, ones that are valid and can immediately be used.</br></br>
	   This is as opposed to the regular `/payment-methods`, which does not categorize payment methods as valid for use.</br></br>
	   You should trigger this when the user is approaching on a gas station with fueling support to get a list of available payment methods.</br></br>
	   If the list is empty, you can ask the user to add a payment method to use PACE fueling.
	*/
	GetPaymentMethodsIncludingCreditCheck(context.Context, GetPaymentMethodsIncludingCreditCheckResponseWriter, *GetPaymentMethodsIncludingCreditCheckRequest) error
	/*
	   GetPaymentMethodsIncludingPaymentToken Get all payment methods with pre-authorized amounts

	   This request returns all payment methods with pre-authorized amounts.</br></br>
	   The list will contain the pre-authorized amount (incl. currency), all information about the payment method and the paymentToken that can be used to complete the payment.</br></br>
	   Empty list if there are no pre-authorized amounts.
	*/
	GetPaymentMethodsIncludingPaymentToken(context.Context, GetPaymentMethodsIncludingPaymentTokenResponseWriter, *GetPaymentMethodsIncludingPaymentTokenRequest) error
}

/*
Router implements: PACE Payment API

Welcome to the PACE Payment API documentation.
This API is responsible for managing payment methods for users as well as authorizing payments on behalf of PACE services.
*/
func Router(service Service) *mux.Router {
	router := mux.NewRouter()
	// Subrouter s1 - https://api.pace.cloud/pay
	s1 := router.PathPrefix("/pay").Subrouter()
	s1.Methods("DELETE").Path("/beta/payment-methods/{paymentMethodId}/paymentTokens/{paymentTokenId}").Handler(DeletePaymentTokenHandler(service)).Name("DeletePaymentToken")
	s1.Methods("POST").Path("/beta/payment-methods/{paymentMethodId}/authorize").Handler(AuthorizePaymentMethodHandler(service)).Name("AuthorizePaymentMethod")
	s1.Methods("POST").Path("/beta/payment-methods/sepa-direct-debit").Handler(CreatePaymentMethodSEPAHandler(service)).Name("CreatePaymentMethodSEPA")
	s1.Methods("DELETE").Path("/beta/payment-methods/{paymentMethodId}").Handler(DeletePaymentMethodHandler(service)).Name("DeletePaymentMethod")
	s1.Methods("GET").Path("/beta/payment-methods").Handler(GetPaymentMethodsIncludingPaymentTokenHandler(service)).Queries("include", "paymentTokens").Name("GetPaymentMethodsIncludingPaymentToken")
	s1.Methods("GET").Path("/beta/payment-methods").Handler(GetPaymentMethodsIncludingCreditCheckHandler(service)).Queries("include", "creditCheck").Name("GetPaymentMethodsIncludingCreditCheck")
	s1.Methods("GET").Path("/beta/payment-methods").Handler(GetPaymentMethodsHandler(service)).Name("GetPaymentMethods")
	return router
}
