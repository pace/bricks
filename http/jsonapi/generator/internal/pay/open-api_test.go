package pay

import (
	"context"
	mux "github.com/gorilla/mux"
	runtime "lab.jamit.de/pace/web/libs/go-microservice/http/jsonapi/runtime"
	"net/http"
)

// AllPaymentMethods ...
type AllPaymentMethods []struct {
	ID                   string `jsonapi:"primary,paymentMethod,omitempty" valid:"uuid,optional"` // Payment method ID
	IdentificationString string `jsonapi:"attr,identificationString,omitempty" valid:"optional"`  // Example: "DE89 **** 3000"
	Kind                 string `jsonapi:"attr,kind,omitempty" valid:"optional,in(sepa)"`         // Example: "sepa"
}

// PaymentMethodSEPA ...
type PaymentMethodSEPA struct {
	ID        string                    `jsonapi:"primary,paymentMethod,omitempty" valid:"uuid,optional"` // The ID of this payment method.
	Address   *PaymentMethodSEPAAddress `jsonapi:"attr,address,omitempty" valid:"required"`
	FirstName string                    `jsonapi:"attr,firstName,omitempty" valid:"required"` // Example: "Jon"
	Iban      string                    `jsonapi:"attr,iban,omitempty" valid:"required"`      // Example: "DE89370400440532013000"
	Kind      string                    `jsonapi:"attr,kind,omitempty" valid:"required,in(sepa)"`
	LastName  string                    `jsonapi:"attr,lastName,omitempty" valid:"required"` // Example: "Smith"
}

// PaymentMethodSEPAAddress ...
type PaymentMethodSEPAAddress struct {
	City        string `jsonapi:"city,omitempty" valid:"required"`        // Example: "Karlsruhe"
	CountryCode string `jsonapi:"countryCode,omitempty" valid:"required"` // Country code in as specified in ISO 3166-1.
	HouseNo     string `jsonapi:"houseNo,omitempty" valid:"required"`     // Example: "18"
	PostalCode  string `jsonapi:"postalCode,omitempty" valid:"required"`  // Example: "76131"
	Street      string `jsonapi:"street,omitempty" valid:"required"`      // Example: "Haid-und-Neu-Str."
}

// PaymentMethodsWithPaymentTokens ...
type PaymentMethodsWithPaymentTokens []struct {
	ID                   string `jsonapi:"primary,paymentMethod,omitempty" valid:"uuid,optional"` // Payment method ID
	IdentificationString string `jsonapi:"attr,identificationString,omitempty" valid:"optional"`  // Example: "DE89 **** 3000"
	Kind                 string `jsonapi:"attr,kind,omitempty" valid:"optional,in(sepa)"`         // Example: "sepa"
}

/*
GetPaymentMethodsHandler handles request/response marshaling and validation for
 Get /payment-methods
*/
func GetPaymentMethodsHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := &getPaymentMethodsResponseWriter{
			ResponseWriter: w,
		}
		request := &GetPaymentMethodsRequest{
			Request: r,
		}
		err := service.GetPaymentMethods(r.Context(), writer, request)
		if err != nil {
			runtime.WriteError(w, http.StatusInternalServerError, err)
		}
	})
}

/*
PostPaymentMethodsPaymentMethodIdAuthorizeHandler handles request/response marshaling and validation for
 Post /payment-methods/:paymentMethodId/authorize
*/
func PostPaymentMethodsPaymentMethodIdAuthorizeHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		writer := &postPaymentMethodsPaymentMethodIdAuthorizeResponseWriter{
			ResponseWriter: w,
		}
		request := &PostPaymentMethodsPaymentMethodIdAuthorizeRequest{
			Request:              r,
			ParamPaymentMethodId: vars["paymentMethodId"],
		}
		if !runtime.ValidateStruct(w, r, &request) {
			return // invalid request stop further processing
		}
		if runtime.Unmarshal(w, r, &request.Content) {
			err := service.PostPaymentMethodsPaymentMethodIdAuthorize(r.Context(), writer, request)
			if err != nil {
				runtime.WriteError(w, http.StatusInternalServerError, err)
			}
		}
	})
}

/*
DeletePaymentMethodsPaymentMethodIdPaymentTokensPaymentTokenIdHandler handles request/response marshaling and validation for
 Delete /payment-methods/:paymentMethodId/paymentTokens/:paymentTokenId
*/
func DeletePaymentMethodsPaymentMethodIdPaymentTokensPaymentTokenIdHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		writer := &deletePaymentMethodsPaymentMethodIdPaymentTokensPaymentTokenIdResponseWriter{
			ResponseWriter: w,
		}
		request := &DeletePaymentMethodsPaymentMethodIdPaymentTokensPaymentTokenIdRequest{
			Request:              r,
			ParamPaymentTokenId:  vars["paymentTokenId"],
			ParamPaymentMethodId: vars["paymentMethodId"],
		}
		if !runtime.ValidateStruct(w, r, &request) {
			return // invalid request stop further processing
		}
		err := service.DeletePaymentMethodsPaymentMethodIdPaymentTokensPaymentTokenId(r.Context(), writer, request)
		if err != nil {
			runtime.WriteError(w, http.StatusInternalServerError, err)
		}
	})
}

/*
PostPaymentMethodsSepaDirectDebitHandler handles request/response marshaling and validation for
 Post /payment-methods/sepa-direct-debit
*/
func PostPaymentMethodsSepaDirectDebitHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := &postPaymentMethodsSepaDirectDebitResponseWriter{
			ResponseWriter: w,
		}
		request := &PostPaymentMethodsSepaDirectDebitRequest{
			Request: r,
		}
		if !runtime.ValidateStruct(w, r, &request) {
			return // invalid request stop further processing
		}
		if runtime.Unmarshal(w, r, &request.Content) {
			err := service.PostPaymentMethodsSepaDirectDebit(r.Context(), writer, request)
			if err != nil {
				runtime.WriteError(w, http.StatusInternalServerError, err)
			}
		}
	})
}

/*
DeletePaymentMethodsPaymentMethodIdHandler handles request/response marshaling and validation for
 Delete /payment-methods/{paymentMethodId}
*/
func DeletePaymentMethodsPaymentMethodIdHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		writer := &deletePaymentMethodsPaymentMethodIdResponseWriter{
			ResponseWriter: w,
		}
		request := &DeletePaymentMethodsPaymentMethodIdRequest{
			Request:              r,
			ParamPaymentMethodId: vars["paymentMethodId"],
		}
		if !runtime.ValidateStruct(w, r, &request) {
			return // invalid request stop further processing
		}
		err := service.DeletePaymentMethodsPaymentMethodId(r.Context(), writer, request)
		if err != nil {
			runtime.WriteError(w, http.StatusInternalServerError, err)
		}
	})
}

/*
GetPaymentMethodsIncludeCreditCheckHandler handles request/response marshaling and validation for
 Get /payment-methods?include=creditCheck
*/
func GetPaymentMethodsIncludeCreditCheckHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		writer := &getPaymentMethodsIncludeCreditCheckResponseWriter{
			ResponseWriter: w,
		}
		request := &GetPaymentMethodsIncludeCreditCheckRequest{
			Request:      r,
			ParamInclude: vars["include"],
		}
		if !runtime.ValidateStruct(w, r, &request) {
			return // invalid request stop further processing
		}
		err := service.GetPaymentMethodsIncludeCreditCheck(r.Context(), writer, request)
		if err != nil {
			runtime.WriteError(w, http.StatusInternalServerError, err)
		}
	})
}

/*
GetPaymentMethodsIncludePaymentTokensHandler handles request/response marshaling and validation for
 Get /payment-methods?include=paymentTokens
*/
func GetPaymentMethodsIncludePaymentTokensHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		writer := &getPaymentMethodsIncludePaymentTokensResponseWriter{
			ResponseWriter: w,
		}
		request := &GetPaymentMethodsIncludePaymentTokensRequest{
			Request:      r,
			ParamInclude: vars["include"],
		}
		if !runtime.ValidateStruct(w, r, &request) {
			return // invalid request stop further processing
		}
		err := service.GetPaymentMethodsIncludePaymentTokens(r.Context(), writer, request)
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
	AllThePaymentMethodsForUser(*AllPaymentMethods)
}
type getPaymentMethodsResponseWriter struct {
	http.ResponseWriter
}

// AllThePaymentMethodsForUser responds with jsonapi marshaled data (HTTP code 200)
func (w *getPaymentMethodsResponseWriter) AllThePaymentMethodsForUser(data *AllPaymentMethods) {
	runtime.Marshal(w, data, 200)
}

/*
GetPaymentMethodsResponseWriter is a standard http.Request extended with the
un-marshaled content object
*/
type GetPaymentMethodsRequest struct {
	Request *http.Request `valid:"-"`
}

// PostPaymentMethodsPaymentMethodIdAuthorizeOK ...
type PostPaymentMethodsPaymentMethodIdAuthorizeOK struct {
	ID       string  `jsonapi:"primary,paymentToken,omitempty" valid:"optional"` // paymentToken ID (NOT the token value)
	Amount   float64 `jsonapi:"attr,amount,omitempty" valid:"optional"`          // Example: "65.49"
	Currency string  `jsonapi:"attr,currency,omitempty" valid:"optional"`        // Currency as specified in ISO-4217.
	Value    string  `jsonapi:"attr,value,omitempty" valid:"optional"`           // The actual token value. Note that the format is subject to change. Treat transparently.
}

/*
PostPaymentMethodsPaymentMethodIdAuthorizeResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type PostPaymentMethodsPaymentMethodIdAuthorizeResponseWriter interface {
	http.ResponseWriter
	OK(*PostPaymentMethodsPaymentMethodIdAuthorizeOK)
	Forbidden(error)
	NotFound(error)
	BadGateway(error)
}
type postPaymentMethodsPaymentMethodIdAuthorizeResponseWriter struct {
	http.ResponseWriter
}

// BadGateway responds with jsonapi error (HTTP code 502)
func (w *postPaymentMethodsPaymentMethodIdAuthorizeResponseWriter) BadGateway(err error) {
	runtime.WriteError(w, 502, err)
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *postPaymentMethodsPaymentMethodIdAuthorizeResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// Forbidden responds with jsonapi error (HTTP code 403)
func (w *postPaymentMethodsPaymentMethodIdAuthorizeResponseWriter) Forbidden(err error) {
	runtime.WriteError(w, 403, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *postPaymentMethodsPaymentMethodIdAuthorizeResponseWriter) OK(data *PostPaymentMethodsPaymentMethodIdAuthorizeOK) {
	runtime.Marshal(w, data, 200)
}

// PostPaymentMethodsPaymentMethodIdAuthorizeContent ...
type PostPaymentMethodsPaymentMethodIdAuthorizeContent struct {
	ID       string  `jsonapi:"primary,paymentToken,omitempty" valid:"uuid,optional"` // ID of the new paymentToken.
	Amount   float64 `jsonapi:"attr,amount,omitempty" valid:"required"`               // Example: "65.49"
	Currency string  `jsonapi:"attr,currency,omitempty" valid:"required"`             // Currency as specified in ISO-4217.
}

// PostPaymentMethodsPaymentMethodIdAuthorizeRequest ...
type PostPaymentMethodsPaymentMethodIdAuthorizeRequest struct {
	Request              *http.Request                                      `valid:"-"`
	Content              *PostPaymentMethodsPaymentMethodIdAuthorizeContent `valid:"-"`
	ParamPaymentMethodId string                                             `valid:"required,uuid"`
}

/*
DeletePaymentMethodsPaymentMethodIdPaymentTokensPaymentTokenIdResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type DeletePaymentMethodsPaymentMethodIdPaymentTokensPaymentTokenIdResponseWriter interface {
	http.ResponseWriter
	ThePaymentTokenWasRemovedSuccessfully()
	NotFound(error)
}
type deletePaymentMethodsPaymentMethodIdPaymentTokensPaymentTokenIdResponseWriter struct {
	http.ResponseWriter
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *deletePaymentMethodsPaymentMethodIdPaymentTokensPaymentTokenIdResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// ThePaymentTokenWasRemovedSuccessfully responds with empty response (HTTP code 204)
func (w *deletePaymentMethodsPaymentMethodIdPaymentTokensPaymentTokenIdResponseWriter) ThePaymentTokenWasRemovedSuccessfully() {
	w.WriteHeader(204)
}

/*
DeletePaymentMethodsPaymentMethodIdPaymentTokensPaymentTokenIdResponseWriter is a standard http.Request extended with the
un-marshaled content object
*/
type DeletePaymentMethodsPaymentMethodIdPaymentTokensPaymentTokenIdRequest struct {
	Request              *http.Request `valid:"-"`
	ParamPaymentTokenId  string        `valid:"required"`
	ParamPaymentMethodId string        `valid:"required,uuid"`
}

// PostPaymentMethodsSepaDirectDebitCreated ...
type PostPaymentMethodsSepaDirectDebitCreated struct {
	ID                   string `jsonapi:"primary,paymentMethod,omitempty" valid:"uuid,optional"` // Payment method ID
	IdentificationString string `jsonapi:"attr,identificationString,omitempty" valid:"optional"`  // Example: "DE89 **** 3000"
	Kind                 string `jsonapi:"attr,kind,omitempty" valid:"optional,in(sepa)"`
}

/*
PostPaymentMethodsSepaDirectDebitResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type PostPaymentMethodsSepaDirectDebitResponseWriter interface {
	http.ResponseWriter
	Created(*PostPaymentMethodsSepaDirectDebitCreated)
	BadRequest(error)
}
type postPaymentMethodsSepaDirectDebitResponseWriter struct {
	http.ResponseWriter
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *postPaymentMethodsSepaDirectDebitResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// Created responds with jsonapi marshaled data (HTTP code 201)
func (w *postPaymentMethodsSepaDirectDebitResponseWriter) Created(data *PostPaymentMethodsSepaDirectDebitCreated) {
	runtime.Marshal(w, data, 201)
}

// PostPaymentMethodsSepaDirectDebitRequest ...
type PostPaymentMethodsSepaDirectDebitRequest struct {
	Request *http.Request      `valid:"-"`
	Content *PaymentMethodSEPA `valid:"-"`
}

/*
DeletePaymentMethodsPaymentMethodIdResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type DeletePaymentMethodsPaymentMethodIdResponseWriter interface {
	http.ResponseWriter
	ThePaymentMethodWasDeletedSuccessfully()
}
type deletePaymentMethodsPaymentMethodIdResponseWriter struct {
	http.ResponseWriter
}

// ThePaymentMethodWasDeletedSuccessfully responds with empty response (HTTP code 204)
func (w *deletePaymentMethodsPaymentMethodIdResponseWriter) ThePaymentMethodWasDeletedSuccessfully() {
	w.WriteHeader(204)
}

/*
DeletePaymentMethodsPaymentMethodIdResponseWriter is a standard http.Request extended with the
un-marshaled content object
*/
type DeletePaymentMethodsPaymentMethodIdRequest struct {
	Request              *http.Request `valid:"-"`
	ParamPaymentMethodId string        `valid:"required,uuid"`
}

/*
GetPaymentMethodsIncludeCreditCheckResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetPaymentMethodsIncludeCreditCheckResponseWriter interface {
	http.ResponseWriter
	AllThePaymentMethodsThatCouldBeUsed(*AllPaymentMethods)
}
type getPaymentMethodsIncludeCreditCheckResponseWriter struct {
	http.ResponseWriter
}

// AllThePaymentMethodsThatCouldBeUsed responds with jsonapi marshaled data (HTTP code 200)
func (w *getPaymentMethodsIncludeCreditCheckResponseWriter) AllThePaymentMethodsThatCouldBeUsed(data *AllPaymentMethods) {
	runtime.Marshal(w, data, 200)
}

/*
GetPaymentMethodsIncludeCreditCheckResponseWriter is a standard http.Request extended with the
un-marshaled content object
*/
type GetPaymentMethodsIncludeCreditCheckRequest struct {
	Request      *http.Request `valid:"-"`
	ParamInclude string        `valid:"required,in(creditCheck)"`
}

/*
GetPaymentMethodsIncludePaymentTokensResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetPaymentMethodsIncludePaymentTokensResponseWriter interface {
	http.ResponseWriter
	AllThePaymentMethodsWithPreAuthorisedAmounts(*PaymentMethodsWithPaymentTokens)
}
type getPaymentMethodsIncludePaymentTokensResponseWriter struct {
	http.ResponseWriter
}

// AllThePaymentMethodsWithPreAuthorisedAmounts responds with jsonapi marshaled data (HTTP code 200)
func (w *getPaymentMethodsIncludePaymentTokensResponseWriter) AllThePaymentMethodsWithPreAuthorisedAmounts(data *PaymentMethodsWithPaymentTokens) {
	runtime.Marshal(w, data, 200)
}

/*
GetPaymentMethodsIncludePaymentTokensResponseWriter is a standard http.Request extended with the
un-marshaled content object
*/
type GetPaymentMethodsIncludePaymentTokensRequest struct {
	Request      *http.Request `valid:"-"`
	ParamInclude string        `valid:"required,in(paymentToken)"`
}
type Service interface {
	// GetPaymentMethods Get all payment methods for user
	GetPaymentMethods(context.Context, GetPaymentMethodsResponseWriter, *GetPaymentMethodsRequest) error
	/*
	   PostPaymentMethodsPaymentMethodIdAuthorize Authorize a payment using the payment method whose ID is paymentMethodId

	   When successful, returns a paymentToken value.
	*/
	PostPaymentMethodsPaymentMethodIdAuthorize(context.Context, PostPaymentMethodsPaymentMethodIdAuthorizeResponseWriter, *PostPaymentMethodsPaymentMethodIdAuthorizeRequest) error
	// DeletePaymentMethodsPaymentMethodIdPaymentTokensPaymentTokenId Delete the paymentToken record.
	DeletePaymentMethodsPaymentMethodIdPaymentTokensPaymentTokenId(context.Context, DeletePaymentMethodsPaymentMethodIdPaymentTokensPaymentTokenIdResponseWriter, *DeletePaymentMethodsPaymentMethodIdPaymentTokensPaymentTokenIdRequest) error
	/*
	   PostPaymentMethodsSepaDirectDebit Register SEPA direct debit as a payment method

	   By registering you allow the user to use SEPA direct debit as a payment method.
	   The payment method ID is optional when posting data.
	*/
	PostPaymentMethodsSepaDirectDebit(context.Context, PostPaymentMethodsSepaDirectDebitResponseWriter, *PostPaymentMethodsSepaDirectDebitRequest) error
	// DeletePaymentMethodsPaymentMethodId Delete a payment method
	DeletePaymentMethodsPaymentMethodId(context.Context, DeletePaymentMethodsPaymentMethodIdResponseWriter, *DeletePaymentMethodsPaymentMethodIdRequest) error
	/*
	   GetPaymentMethodsIncludeCreditCheck Get all ready-to-use payment methods for user

	   This request will return a list of supported payment methods for the current user that they can, in theory, use. That is, ones that are valid and can immediately be used.</br></br>
	   This is as opposed to the regular `/payment-methods`, which does not categorize payment methods as valid for use.</br></br>
	   You should trigger this when the user is approaching on a gas station with fueling support to get a list of available payment methods.</br></br>
	   If the list is empty, you can ask the user to add a payment method to use PACE fueling.
	*/
	GetPaymentMethodsIncludeCreditCheck(context.Context, GetPaymentMethodsIncludeCreditCheckResponseWriter, *GetPaymentMethodsIncludeCreditCheckRequest) error
	/*
	   GetPaymentMethodsIncludePaymentTokens Get all payment methods with pre-authorized amounts

	   This request returns all payment methods with pre-authorized amounts.</br></br>
	   The list will contain the pre-authorized amount (incl. currency), all information about the payment method and the paymentToken that can be used to complete the payment.</br></br>
	   Empty list if there are no pre-authorized amounts.
	*/
	GetPaymentMethodsIncludePaymentTokens(context.Context, GetPaymentMethodsIncludePaymentTokensResponseWriter, *GetPaymentMethodsIncludePaymentTokensRequest) error
}

/*
Router implements: PACE Payment API

Welcome to the PACE Payment API documentation.
This API is responsible for managing payment methods for users as well as authorizing payments on behalf of PACE services.
*/
func Router(service Service) *mux.Router {
	router := mux.NewRouter()
	// Subrouter s1 - https://api.pace.cloud/pay/beta
	s1 := router.PathPrefix("/pay/beta").Subrouter()
	s1.Methods("DELETE").Path("/payment-methods/:paymentMethodId/paymentTokens/:paymentTokenId").Handler(DeletePaymentMethodsPaymentMethodIdPaymentTokensPaymentTokenIdHandler(service)).Name("DeletePaymentMethodsPaymentMethodIdPaymentTokensPaymentTokenId")
	s1.Methods("POST").Path("/payment-methods/:paymentMethodId/authorize").Handler(PostPaymentMethodsPaymentMethodIdAuthorizeHandler(service)).Name("PostPaymentMethodsPaymentMethodIdAuthorize")
	s1.Methods("POST").Path("/payment-methods/sepa-direct-debit").Handler(PostPaymentMethodsSepaDirectDebitHandler(service)).Name("PostPaymentMethodsSepaDirectDebit")
	s1.Methods("DELETE").Path("/payment-methods/{paymentMethodId}").Handler(DeletePaymentMethodsPaymentMethodIdHandler(service)).Name("DeletePaymentMethodsPaymentMethodId")
	s1.Methods("GET").Path("/payment-methods").Handler(GetPaymentMethodsIncludePaymentTokensHandler(service)).Queries("include", "paymentTokens").Name("GetPaymentMethodsIncludePaymentTokens")
	s1.Methods("GET").Path("/payment-methods").Handler(GetPaymentMethodsIncludeCreditCheckHandler(service)).Queries("include", "creditCheck").Name("GetPaymentMethodsIncludeCreditCheck")
	s1.Methods("GET").Path("/payment-methods").Handler(GetPaymentMethodsHandler(service)).Name("GetPaymentMethods")
	return router
}
