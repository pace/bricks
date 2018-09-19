package pay

import (
	"context"
	"errors"
	mux "github.com/gorilla/mux"
	opentracing "github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
	runtime "lab.jamit.de/pace/go-microservice/http/jsonapi/runtime"
	log "lab.jamit.de/pace/go-microservice/maintenance/log"
	jsonapimetrics "lab.jamit.de/pace/go-microservice/maintenance/metrics/jsonapi"
	oauth2 "lab.jamit.de/pace/go-microservice/oauth2"
	"net/http"
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
	ID                   string          `jsonapi:"primary,paymentMethod,omitempty" valid:"uuid,optional"`                                      // Payment method ID
	IdentificationString string          `json:"identificationString,omitempty" jsonapi:"attr,identificationString,omitempty" valid:"optional"` // Example: "DE89 **** 3000"
	Kind                 string          `json:"kind,omitempty" jsonapi:"attr,kind,omitempty" valid:"optional"`                                 // Example: "sepa"
	PaymentTokens        []*PaymentToken `json:"paymentTokens,omitempty" jsonapi:"attr,paymentTokens,omitempty" valid:"optional"`
}

// PaymentMethodsWithPaymentTokens ...
type PaymentMethodsWithPaymentTokens []*PaymentMethodsWithPaymentTokensItem

// PaymentToken ...
type PaymentToken struct {
	ID string `jsonapi:"primary,paymentToken,omitempty" valid:"optional"` // Payment Token ID (externally provided - by payment provider)
}

// TransactionRequestFueling ...
type TransactionRequestFueling struct {
	AppID   string `json:"appId,omitempty" jsonapi:"appId,omitempty" valid:"required"`     // Location-based App ID
	Mileage int64  `json:"mileage,omitempty" jsonapi:"mileage,omitempty" valid:"required"` // Current mileage in meters
	PumpID  string `json:"pumpId,omitempty" jsonapi:"pumpId,omitempty" valid:"required"`   // Pump ID
	Vin     string `json:"vin,omitempty" jsonapi:"vin,omitempty" valid:"required"`         // Example: "1B3EL46R36N102271"
}

// TransactionRequest ...
type TransactionRequest struct {
	ID                string                     `jsonapi:"primary,transaction,omitempty" valid:"uuid,optional"` // Transaction ID
	Currency          *Currency                  `json:"currency,omitempty" jsonapi:"attr,currency,omitempty" valid:"optional"`
	Fueling           *TransactionRequestFueling `json:"fueling,omitempty" jsonapi:"attr,fueling,omitempty" valid:"optional"`
	PaymentToken      string                     `json:"paymentToken,omitempty" jsonapi:"attr,paymentToken,omitempty" valid:"required"`           // Example: "f106ac99-213c-4cf7-8c1b-1e841516026b"
	PriceIncludingVAT float32                    `json:"priceIncludingVAT,omitempty" jsonapi:"attr,priceIncludingVAT,omitempty" valid:"optional"` // Example: "69.34"
}

// Currency ...
type Currency string

/*
GetPaymentMethodsHandler handles request/response marshaling and validation for
 Get /beta/payment-methods
*/
func GetPaymentMethodsHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer func() {
			if rp := recover(); rp != nil {
				log.Ctx(ctx).Error().Str("handler", "GetPaymentMethodsHandler").Msgf("Panic: %v", rp)
				log.Stack(ctx)
				runtime.WriteError(w, http.StatusInternalServerError, errors.New("Error"))
			}
		}()

		// Trace the service function handler execution
		var handlerSpan opentracing.Span
		wireContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("Couldn't get span from request header")
		}
		handlerSpan = opentracing.StartSpan("GetPaymentMethodsHandler", opentracing.ChildOf(wireContext))
		handlerSpan.LogFields(olog.String("req_id", log.RequestID(r)), olog.String("client_id", oauth2.ClientID(r.Context())), olog.String("user_id", oauth2.UserID(r.Context())))
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		ctx = opentracing.ContextWithSpan(r.Context(), handlerSpan)
		writer := getPaymentMethodsResponseWriter{
			ResponseWriter: jsonapimetrics.NewMetric("pay", "/beta/payment-methods", w, r),
		}
		request := GetPaymentMethodsRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters

		// Invoke service that implements the business logic
		err = service.GetPaymentMethods(ctx, &writer, &request)
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
		ctx := r.Context()
		defer func() {
			if rp := recover(); rp != nil {
				log.Ctx(ctx).Error().Str("handler", "CreatePaymentMethodSEPAHandler").Msgf("Panic: %v", rp)
				log.Stack(ctx)
				runtime.WriteError(w, http.StatusInternalServerError, errors.New("Error"))
			}
		}()

		// Trace the service function handler execution
		var handlerSpan opentracing.Span
		wireContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("Couldn't get span from request header")
		}
		handlerSpan = opentracing.StartSpan("CreatePaymentMethodSEPAHandler", opentracing.ChildOf(wireContext))
		handlerSpan.LogFields(olog.String("req_id", log.RequestID(r)), olog.String("client_id", oauth2.ClientID(r.Context())), olog.String("user_id", oauth2.UserID(r.Context())))
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		ctx = opentracing.ContextWithSpan(r.Context(), handlerSpan)
		writer := createPaymentMethodSEPAResponseWriter{
			ResponseWriter: jsonapimetrics.NewMetric("pay", "/beta/payment-methods/sepa-direct-debit", w, r),
		}
		request := CreatePaymentMethodSEPARequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Unmarshal the service request body
		if runtime.Unmarshal(w, r, &request.Content) {
			// Invoke service that implements the business logic
			err = service.CreatePaymentMethodSEPA(ctx, &writer, &request)
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
		ctx := r.Context()
		defer func() {
			if rp := recover(); rp != nil {
				log.Ctx(ctx).Error().Str("handler", "DeletePaymentMethodHandler").Msgf("Panic: %v", rp)
				log.Stack(ctx)
				runtime.WriteError(w, http.StatusInternalServerError, errors.New("Error"))
			}
		}()

		// Trace the service function handler execution
		var handlerSpan opentracing.Span
		wireContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("Couldn't get span from request header")
		}
		handlerSpan = opentracing.StartSpan("DeletePaymentMethodHandler", opentracing.ChildOf(wireContext))
		handlerSpan.LogFields(olog.String("req_id", log.RequestID(r)), olog.String("client_id", oauth2.ClientID(r.Context())), olog.String("user_id", oauth2.UserID(r.Context())))
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		ctx = opentracing.ContextWithSpan(r.Context(), handlerSpan)
		writer := deletePaymentMethodResponseWriter{
			ResponseWriter: jsonapimetrics.NewMetric("pay", "/beta/payment-methods/{paymentMethodId}", w, r),
		}
		request := DeletePaymentMethodRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
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

		// Invoke service that implements the business logic
		err = service.DeletePaymentMethod(ctx, &writer, &request)
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
		ctx := r.Context()
		defer func() {
			if rp := recover(); rp != nil {
				log.Ctx(ctx).Error().Str("handler", "AuthorizePaymentMethodHandler").Msgf("Panic: %v", rp)
				log.Stack(ctx)
				runtime.WriteError(w, http.StatusInternalServerError, errors.New("Error"))
			}
		}()

		// Trace the service function handler execution
		var handlerSpan opentracing.Span
		wireContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("Couldn't get span from request header")
		}
		handlerSpan = opentracing.StartSpan("AuthorizePaymentMethodHandler", opentracing.ChildOf(wireContext))
		handlerSpan.LogFields(olog.String("req_id", log.RequestID(r)), olog.String("client_id", oauth2.ClientID(r.Context())), olog.String("user_id", oauth2.UserID(r.Context())))
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		ctx = opentracing.ContextWithSpan(r.Context(), handlerSpan)
		writer := authorizePaymentMethodResponseWriter{
			ResponseWriter: jsonapimetrics.NewMetric("pay", "/beta/payment-methods/{paymentMethodId}/authorize", w, r),
		}
		request := AuthorizePaymentMethodRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
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

		// Unmarshal the service request body
		if runtime.Unmarshal(w, r, &request.Content) {
			// Invoke service that implements the business logic
			err = service.AuthorizePaymentMethod(ctx, &writer, &request)
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
		ctx := r.Context()
		defer func() {
			if rp := recover(); rp != nil {
				log.Ctx(ctx).Error().Str("handler", "DeletePaymentTokenHandler").Msgf("Panic: %v", rp)
				log.Stack(ctx)
				runtime.WriteError(w, http.StatusInternalServerError, errors.New("Error"))
			}
		}()

		// Trace the service function handler execution
		var handlerSpan opentracing.Span
		wireContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("Couldn't get span from request header")
		}
		handlerSpan = opentracing.StartSpan("DeletePaymentTokenHandler", opentracing.ChildOf(wireContext))
		handlerSpan.LogFields(olog.String("req_id", log.RequestID(r)), olog.String("client_id", oauth2.ClientID(r.Context())), olog.String("user_id", oauth2.UserID(r.Context())))
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		ctx = opentracing.ContextWithSpan(r.Context(), handlerSpan)
		writer := deletePaymentTokenResponseWriter{
			ResponseWriter: jsonapimetrics.NewMetric("pay", "/beta/payment-methods/{paymentMethodId}/paymentTokens/{paymentTokenId}", w, r),
		}
		request := DeletePaymentTokenRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
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

		// Invoke service that implements the business logic
		err = service.DeletePaymentToken(ctx, &writer, &request)
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
		ctx := r.Context()
		defer func() {
			if rp := recover(); rp != nil {
				log.Ctx(ctx).Error().Str("handler", "GetPaymentMethodsIncludingCreditCheckHandler").Msgf("Panic: %v", rp)
				log.Stack(ctx)
				runtime.WriteError(w, http.StatusInternalServerError, errors.New("Error"))
			}
		}()

		// Trace the service function handler execution
		var handlerSpan opentracing.Span
		wireContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("Couldn't get span from request header")
		}
		handlerSpan = opentracing.StartSpan("GetPaymentMethodsIncludingCreditCheckHandler", opentracing.ChildOf(wireContext))
		handlerSpan.LogFields(olog.String("req_id", log.RequestID(r)), olog.String("client_id", oauth2.ClientID(r.Context())), olog.String("user_id", oauth2.UserID(r.Context())))
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		ctx = opentracing.ContextWithSpan(r.Context(), handlerSpan)
		writer := getPaymentMethodsIncludingCreditCheckResponseWriter{
			ResponseWriter: jsonapimetrics.NewMetric("pay", "/beta/payment-methods?include=creditCheck", w, r),
		}
		request := GetPaymentMethodsIncludingCreditCheckRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
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

		// Invoke service that implements the business logic
		err = service.GetPaymentMethodsIncludingCreditCheck(ctx, &writer, &request)
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
		ctx := r.Context()
		defer func() {
			if rp := recover(); rp != nil {
				log.Ctx(ctx).Error().Str("handler", "GetPaymentMethodsIncludingPaymentTokenHandler").Msgf("Panic: %v", rp)
				log.Stack(ctx)
				runtime.WriteError(w, http.StatusInternalServerError, errors.New("Error"))
			}
		}()

		// Trace the service function handler execution
		var handlerSpan opentracing.Span
		wireContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("Couldn't get span from request header")
		}
		handlerSpan = opentracing.StartSpan("GetPaymentMethodsIncludingPaymentTokenHandler", opentracing.ChildOf(wireContext))
		handlerSpan.LogFields(olog.String("req_id", log.RequestID(r)), olog.String("client_id", oauth2.ClientID(r.Context())), olog.String("user_id", oauth2.UserID(r.Context())))
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		ctx = opentracing.ContextWithSpan(r.Context(), handlerSpan)
		writer := getPaymentMethodsIncludingPaymentTokenResponseWriter{
			ResponseWriter: jsonapimetrics.NewMetric("pay", "/beta/payment-methods?include=paymentTokens", w, r),
		}
		request := GetPaymentMethodsIncludingPaymentTokenRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
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

		// Invoke service that implements the business logic
		err = service.GetPaymentMethodsIncludingPaymentToken(ctx, &writer, &request)
		if err != nil {
			runtime.WriteError(w, http.StatusInternalServerError, err)
		}
	})
}

/*
ProcessPaymentHandler handles request/response marshaling and validation for
 Post /beta/transaction
*/
func ProcessPaymentHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer func() {
			if rp := recover(); rp != nil {
				log.Ctx(ctx).Error().Str("handler", "ProcessPaymentHandler").Msgf("Panic: %v", rp)
				log.Stack(ctx)
				runtime.WriteError(w, http.StatusInternalServerError, errors.New("Error"))
			}
		}()

		// Trace the service function handler execution
		var handlerSpan opentracing.Span
		wireContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("Couldn't get span from request header")
		}
		handlerSpan = opentracing.StartSpan("ProcessPaymentHandler", opentracing.ChildOf(wireContext))
		handlerSpan.LogFields(olog.String("req_id", log.RequestID(r)), olog.String("client_id", oauth2.ClientID(r.Context())), olog.String("user_id", oauth2.UserID(r.Context())))
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		ctx = opentracing.ContextWithSpan(r.Context(), handlerSpan)
		writer := processPaymentResponseWriter{
			ResponseWriter: jsonapimetrics.NewMetric("pay", "/beta/transaction", w, r),
		}
		request := ProcessPaymentRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Unmarshal the service request body
		if runtime.Unmarshal(w, r, &request.Content) {
			// Invoke service that implements the business logic
			err = service.ProcessPayment(ctx, &writer, &request)
			if err != nil {
				runtime.WriteError(w, http.StatusInternalServerError, err)
			}
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
GetPaymentMethodsRequest is a standard http.Request extended with the
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
DeletePaymentMethodRequest is a standard http.Request extended with the
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
DeletePaymentTokenRequest is a standard http.Request extended with the
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
GetPaymentMethodsIncludingCreditCheckRequest is a standard http.Request extended with the
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
GetPaymentMethodsIncludingPaymentTokenRequest is a standard http.Request extended with the
un-marshaled content object
*/
type GetPaymentMethodsIncludingPaymentTokenRequest struct {
	Request      *http.Request `valid:"-"`
	ParamInclude string        `valid:"required,in(paymentToken)"`
}

// ProcessPaymentCreated ...
type ProcessPaymentCreated struct {
	ID                string                        `jsonapi:"primary,transaction,omitempty" valid:"uuid,optional"` // Transaction ID
	VAT               *ProcessPaymentCreatedVAT     `json:"VAT,omitempty" jsonapi:"attr,VAT,omitempty" valid:"optional"`
	Currency          *Currency                     `json:"currency,omitempty" jsonapi:"attr,currency,omitempty" valid:"optional"`
	Fueling           *ProcessPaymentCreatedFueling `json:"fueling,omitempty" jsonapi:"attr,fueling,omitempty" valid:"optional"`
	PaymentToken      string                        `json:"paymentToken,omitempty" jsonapi:"attr,paymentToken,omitempty" valid:"optional"`           // Example: "f106ac99-213c-4cf7-8c1b-1e841516026b"
	PriceIncludingVAT float32                       `json:"priceIncludingVAT,omitempty" jsonapi:"attr,priceIncludingVAT,omitempty" valid:"optional"` // Example: "69.34"
	PriceWithoutVAT   float32                       `json:"priceWithoutVAT,omitempty" jsonapi:"attr,priceWithoutVAT,omitempty" valid:"optional"`     // Example: "58.27"
}

// ProcessPaymentCreatedVAT ...
type ProcessPaymentCreatedVAT struct {
	Amount float32 `json:"amount,omitempty" jsonapi:"amount,omitempty" valid:"optional"` // Example: "11.07"
	Rate   float32 `json:"rate,omitempty" jsonapi:"rate,omitempty" valid:"optional"`     // Example: "0.19"
}

// ProcessPaymentCreatedFueling ...
type ProcessPaymentCreatedFueling struct {
	AppID   string `json:"appId,omitempty" jsonapi:"appId,omitempty" valid:"required"`     // Example: "c30bce97-b732-4390-af38-1ac6b017aa4c"
	Mileage int64  `json:"mileage,omitempty" jsonapi:"mileage,omitempty" valid:"required"` // Example: "66435"
	PumpID  string `json:"pumpId,omitempty" jsonapi:"pumpId,omitempty" valid:"required"`   // Example: "460ffaad-a3c1-4199-b69e-63949ccda82f"
	Vin     string `json:"vin,omitempty" jsonapi:"vin,omitempty" valid:"required"`         // Example: "1B3EL46R36N102271"
}

/*
ProcessPaymentResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type ProcessPaymentResponseWriter interface {
	http.ResponseWriter
	Created(*ProcessPaymentCreated)
	BadRequest(error)
	NotFound(error)
	Conflict(error)
}
type processPaymentResponseWriter struct {
	http.ResponseWriter
}

// Conflict responds with jsonapi error (HTTP code 409)
func (w *processPaymentResponseWriter) Conflict(err error) {
	runtime.WriteError(w, 409, err)
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *processPaymentResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *processPaymentResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// Created responds with jsonapi marshaled data (HTTP code 201)
func (w *processPaymentResponseWriter) Created(data *ProcessPaymentCreated) {
	runtime.Marshal(w, data, 201)
}

// ProcessPaymentRequest ...
type ProcessPaymentRequest struct {
	Request *http.Request      `valid:"-"`
	Content TransactionRequest `valid:"-"`
}

// Service interface for all handlers
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
	/*
	   ProcessPayment Process payment

	   Process payment and notify user if transaction is finished successfully. You can optionally provide `priceIncludingVAT`and `currency` in the request body to check if the price the user has seen is still correct.
	*/
	ProcessPayment(context.Context, ProcessPaymentResponseWriter, *ProcessPaymentRequest) error
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
	s1.Methods("POST").Path("/beta/transaction").Handler(ProcessPaymentHandler(service)).Name("ProcessPayment")
	return router
}
