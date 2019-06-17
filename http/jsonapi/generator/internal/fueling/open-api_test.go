// nolint
package fueling

import (
	"context"
	mux "github.com/gorilla/mux"
	opentracing "github.com/opentracing/opentracing-go"
	runtime "github.com/pace/bricks/http/jsonapi/runtime"
	errors "github.com/pace/bricks/maintenance/errors"
	metrics "github.com/pace/bricks/maintenance/metric/jsonapi"
	"net/http"
)

// ApproachingRequest ...
type ApproachingRequest struct {
	ID             string  `jsonapi:"primary,approaching,omitempty" valid:"uuid,optional"`                                                                                                                                        // Approaching ID
	CarFuelType    string  `json:"carFuelType,omitempty" jsonapi:"attr,carFuelType,omitempty" valid:"required,in(e85|ron91|ron95_e5|ron95_e10|ron98|ron98_e5|ron100|diesel|diesel_gtl|diesel_b7|lpg|cng|h2|Truck Diesel|AdBlue)"` // Fuel type of the car
	ExpectedAmount float32 `json:"expectedAmount,omitempty" jsonapi:"attr,expectedAmount,omitempty" valid:"required"`                                                                                                             // Expected amount in liters for refuel
}

// ApproachingResponse ...
type ApproachingResponse struct {
	ID             string           `jsonapi:"primary,approaching,omitempty" valid:"uuid,optional"`                                                                                                                                        // Approaching ID
	CarFuelType    string           `json:"carFuelType,omitempty" jsonapi:"attr,carFuelType,omitempty" valid:"optional,in(e85|ron91|ron95_e5|ron95_e10|ron98|ron98_e5|ron100|diesel|diesel_gtl|diesel_b7|lpg|cng|h2|Truck Diesel|AdBlue)"` // Fuel type of the car
	ExpectedAmount float32          `json:"expectedAmount,omitempty" jsonapi:"attr,expectedAmount,omitempty" valid:"optional"`                                                                                                             // Expected amount in liters for refuel
	GasStation     *GasStation      `json:"gasStation,omitempty" jsonapi:"relation,gasStation,omitempty" valid:"optional"`
	PaymentMethods []*PaymentMethod `json:"paymentMethods,omitempty" jsonapi:"relation,paymentMethods,omitempty" valid:"optional"`
}

// FuelPrice ...
type FuelPrice struct {
	ID          string   `jsonapi:"primary,fuelPrice,omitempty" valid:"optional"` // Fuel Price ID
	Currency    Currency `json:"currency,omitempty" jsonapi:"attr,currency,omitempty" valid:"optional"`
	FuelType    string   `json:"fuelType,omitempty" jsonapi:"attr,fuelType,omitempty" valid:"optional,in(e85|ron91|ron95_e5|ron95_e10|ron98|ron98_e5|ron100|diesel|diesel_gtl|diesel_b7|lpg|cng|h2|Truck Diesel|AdBlue)"` // Example: "ron95_e10"
	Price       float32  `json:"price,omitempty" jsonapi:"attr,price,omitempty" valid:"optional"`                                                                                                                         // Price in liters
	ProductName string   `json:"productName,omitempty" jsonapi:"attr,productName,omitempty" valid:"optional"`                                                                                                             // Example: "Super E10"
}

// FuelPriceResponse ...
type FuelPriceResponse *FuelPrice

// GasStationAddress ...
type GasStationAddress struct {
	City        string `json:"city,omitempty" jsonapi:"attr,city,omitempty" valid:"optional"`               // Example: "Karlsruhe"
	CountryCode string `json:"countryCode,omitempty" jsonapi:"attr,countryCode,omitempty" valid:"optional"` // Country code in as specified in ISO 3166-1.
	HouseNo     string `json:"houseNo,omitempty" jsonapi:"attr,houseNo,omitempty" valid:"optional"`         // Example: "18"
	PostalCode  string `json:"postalCode,omitempty" jsonapi:"attr,postalCode,omitempty" valid:"optional"`   // Example: "76131"
	Street      string `json:"street,omitempty" jsonapi:"attr,street,omitempty" valid:"optional"`           // Example: "Haid-und-Neu-Str."
}

// GasStationOpeningHours ...
type GasStationOpeningHours struct {
	OpenFromTo []string `json:"openFromTo,omitempty" jsonapi:"attr,openFromTo,omitempty" valid:"optional"` // Example: "[07:30 20:30]"
	Weekdays   []string `json:"weekdays,omitempty" jsonapi:"attr,weekdays,omitempty" valid:"optional"`     // Example: "[Montag Dienstag]"
}

// GasStation ...
type GasStation struct {
	ID           string                   `jsonapi:"primary,gasStation,omitempty" valid:"uuid,optional"` // Gas Station ID
	Address      GasStationAddress        `json:"address,omitempty" jsonapi:"attr,address,omitempty" valid:"optional"`
	Amenities    []string                 `json:"amenities,omitempty" jsonapi:"attr,amenities,omitempty" valid:"optional"` // Example: "[restaurant]"
	Latitude     float32                  `json:"latitude,omitempty" jsonapi:"attr,latitude,omitempty" valid:"optional"`   // Example: "49.013"
	Longitude    float32                  `json:"longitude,omitempty" jsonapi:"attr,longitude,omitempty" valid:"optional"` // Example: "8.425"
	OpeningHours []GasStationOpeningHours `json:"openingHours,omitempty" jsonapi:"attr,openingHours,omitempty" valid:"optional"`
	StationName  string                   `json:"stationName,omitempty" jsonapi:"attr,stationName,omitempty" valid:"optional"` // Example: "PACE Station"
	FuelPrices   []*FuelPrice             `json:"fuelPrices,omitempty" jsonapi:"relation,fuelPrices,omitempty" valid:"optional"`
	Pumps        []*Pump                  `json:"pumps,omitempty" jsonapi:"relation,pumps,omitempty" valid:"optional"`
}

// PaymentMethod ...
type PaymentMethod struct {
	ID   string `jsonapi:"primary,paymentMethod,omitempty" valid:"optional"`           // Payment Method ID
	Kind string `json:"kind,omitempty" jsonapi:"attr,kind,omitempty" valid:"optional"` // Example: "sepa"
}

// PaymentMethodResponse ...
type PaymentMethodResponse *PaymentMethod

// Pump ...
type Pump struct {
	ID         string     `jsonapi:"primary,pump,omitempty" valid:"uuid,optional"`                           // Pump ID
	Identifier string     `json:"identifier,omitempty" jsonapi:"attr,identifier,omitempty" valid:"optional"` // Pump identifier
	Status     PumpStatus `json:"status,omitempty" jsonapi:"attr,status,omitempty" valid:"optional"`
}

// PumpResponseVAT ...
type PumpResponseVAT struct {
	Amount float32 `json:"amount,omitempty" jsonapi:"attr,amount,omitempty" valid:"optional"` // Example: "9.72"
	Rate   float32 `json:"rate,omitempty" jsonapi:"attr,rate,omitempty" valid:"optional"`     // Example: "0.19"
}

// PumpResponse ...
type PumpResponse struct {
	ID                string          `jsonapi:"primary,pump,omitempty" valid:"uuid,optional"` // Pump ID
	VAT               PumpResponseVAT `json:"VAT,omitempty" jsonapi:"attr,VAT,omitempty" valid:"optional"`
	Currency          Currency        `json:"currency,omitempty" jsonapi:"attr,currency,omitempty" valid:"optional"`
	FuelAmount        float32         `json:"fuelAmount,omitempty" jsonapi:"attr,fuelAmount,omitempty" valid:"optional"`               // Fuel amount in liters
	FuelType          string          `json:"fuelType,omitempty" jsonapi:"attr,fuelType,omitempty" valid:"optional"`                   // Example: "ron95_e10"
	Identifier        string          `json:"identifier,omitempty" jsonapi:"attr,identifier,omitempty" valid:"optional"`               // Pump identifier
	PriceIncludingVAT float32         `json:"priceIncludingVAT,omitempty" jsonapi:"attr,priceIncludingVAT,omitempty" valid:"optional"` // Example: "61.09"
	PriceWithoutVAT   float32         `json:"priceWithoutVAT,omitempty" jsonapi:"attr,priceWithoutVAT,omitempty" valid:"optional"`     // Example: "51.37"
	ProductName       string          `json:"productName,omitempty" jsonapi:"attr,productName,omitempty" valid:"optional"`             // Example: "Super E10"
	Status            PumpStatus      `json:"status,omitempty" jsonapi:"attr,status,omitempty" valid:"optional"`
}

// PumpStatus Current pump status
type PumpStatus string

// TransactionRequest ...
type TransactionRequest struct {
	ID                string   `jsonapi:"primary,transaction,omitempty" valid:"uuid,optional"` // Transaction ID
	Currency          Currency `json:"currency,omitempty" jsonapi:"attr,currency,omitempty" valid:"optional"`
	FuelingAppID      string   `json:"fuelingAppId,omitempty" jsonapi:"attr,fuelingAppId,omitempty" valid:"required,uuid"`      // Location-based App ID
	Mileage           int64    `json:"mileage,omitempty" jsonapi:"attr,mileage,omitempty" valid:"optional"`                     // Current mileage in meters
	PaymentToken      string   `json:"paymentToken,omitempty" jsonapi:"attr,paymentToken,omitempty" valid:"required"`           // Example: "f106ac99-213c-4cf7-8c1b-1e841516026b"
	PriceIncludingVAT float32  `json:"priceIncludingVAT,omitempty" jsonapi:"attr,priceIncludingVAT,omitempty" valid:"optional"` // Example: "69.34"
	PumpID            string   `json:"pumpId,omitempty" jsonapi:"attr,pumpId,omitempty" valid:"required,uuid"`                  // Pump ID
	Vin               string   `json:"vin,omitempty" jsonapi:"attr,vin,omitempty" valid:"optional"`                             // Example: "1B3EL46R36N102271"
}

// Currency ...
type Currency string

/*
ProcessPaymentHandler handles request/response marshaling and validation for
 Post /beta/gas-station/{gasStationId}/payment
*/
func ProcessPaymentHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("ProcessPaymentHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "ProcessPaymentHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := processPaymentResponseWriter{
			ResponseWriter: metrics.NewMetric("fueling", "/beta/gas-station/{gasStationId}/payment", w, r),
		}
		request := ProcessPaymentRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamGasStationID,
			Location: runtime.ScanInPath,
			Input:    vars["gasStationId"],
			Name:     "gasStationId",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Unmarshal the service request body
		if runtime.Unmarshal(w, r, &request.Content) {
			// Invoke service that implements the business logic
			err := service.ProcessPayment(ctx, &writer, &request)
			if err != nil {
				errors.HandleError(err, "ProcessPaymentHandler", w, r)
			}
		}
	})
}

/*
ApproachingAtTheForecourtHandler handles request/response marshaling and validation for
 Post /beta/gas-stations/{gasStationId}/approaching
*/
func ApproachingAtTheForecourtHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("ApproachingAtTheForecourtHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "ApproachingAtTheForecourtHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := approachingAtTheForecourtResponseWriter{
			ResponseWriter: metrics.NewMetric("fueling", "/beta/gas-stations/{gasStationId}/approaching", w, r),
		}
		request := ApproachingAtTheForecourtRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamGasStationID,
			Location: runtime.ScanInPath,
			Input:    vars["gasStationId"],
			Name:     "gasStationId",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Unmarshal the service request body
		if runtime.Unmarshal(w, r, &request.Content) {
			// Invoke service that implements the business logic
			err := service.ApproachingAtTheForecourt(ctx, &writer, &request)
			if err != nil {
				errors.HandleError(err, "ApproachingAtTheForecourtHandler", w, r)
			}
		}
	})
}

/*
GetPumpHandler handles request/response marshaling and validation for
 Get /beta/gas-stations/{gasStationId}/pumps/{pumpId}
*/
func GetPumpHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("GetPumpHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "GetPumpHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := getPumpResponseWriter{
			ResponseWriter: metrics.NewMetric("fueling", "/beta/gas-stations/{gasStationId}/pumps/{pumpId}", w, r),
		}
		request := GetPumpRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamGasStationID,
			Location: runtime.ScanInPath,
			Input:    vars["gasStationId"],
			Name:     "gasStationId",
		}, &runtime.ScanParameter{
			Data:     &request.ParamPumpID,
			Location: runtime.ScanInPath,
			Input:    vars["pumpId"],
			Name:     "pumpId",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Invoke service that implements the business logic
		err := service.GetPump(ctx, &writer, &request)
		if err != nil {
			errors.HandleError(err, "GetPumpHandler", w, r)
		}
	})
}

/*
WaitOnPumpStatusChangeHandler handles request/response marshaling and validation for
 Get /beta/gas-stations/{gasStationId}/pumps/{pumpId}/wait-for-status-change
*/
func WaitOnPumpStatusChangeHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("WaitOnPumpStatusChangeHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "WaitOnPumpStatusChangeHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := waitOnPumpStatusChangeResponseWriter{
			ResponseWriter: metrics.NewMetric("fueling", "/beta/gas-stations/{gasStationId}/pumps/{pumpId}/wait-for-status-change", w, r),
		}
		request := WaitOnPumpStatusChangeRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamGasStationID,
			Location: runtime.ScanInPath,
			Input:    vars["gasStationId"],
			Name:     "gasStationId",
		}, &runtime.ScanParameter{
			Data:     &request.ParamPumpID,
			Location: runtime.ScanInPath,
			Input:    vars["pumpId"],
			Name:     "pumpId",
		}, &runtime.ScanParameter{
			Data:     &request.ParamUpdate,
			Location: runtime.ScanInQuery,
			Input:    vars["update"],
			Name:     "update",
		}, &runtime.ScanParameter{
			Data:     &request.ParamLastStatus,
			Location: runtime.ScanInQuery,
			Input:    vars["lastStatus"],
			Name:     "lastStatus",
		}, &runtime.ScanParameter{
			Data:     &request.ParamTimeout,
			Location: runtime.ScanInQuery,
			Input:    vars["timeout"],
			Name:     "timeout",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Invoke service that implements the business logic
		err := service.WaitOnPumpStatusChange(ctx, &writer, &request)
		if err != nil {
			errors.HandleError(err, "WaitOnPumpStatusChangeHandler", w, r)
		}
	})
}

// ProcessPaymentCreated ...
type ProcessPaymentCreated struct {
	ID                string                   `jsonapi:"primary,transaction,omitempty" valid:"uuid,optional"` // Transaction ID
	VAT               ProcessPaymentCreatedVAT `json:"VAT,omitempty" jsonapi:"attr,VAT,omitempty" valid:"optional"`
	Currency          Currency                 `json:"currency,omitempty" jsonapi:"attr,currency,omitempty" valid:"optional"`
	FuelingAppID      string                   `json:"fuelingAppId,omitempty" jsonapi:"attr,fuelingAppId,omitempty" valid:"optional,uuid"`      // Example: "c30bce97-b732-4390-af38-1ac6b017aa4c"
	GasStationID      string                   `json:"gasStationId,omitempty" jsonapi:"attr,gasStationId,omitempty" valid:"optional,uuid"`      // Example: "a6ec9bd7-cf0b-416c-b24f-9ce65ab3dfe1"
	Mileage           int64                    `json:"mileage,omitempty" jsonapi:"attr,mileage,omitempty" valid:"optional"`                     // Example: "66435"
	PaymentToken      string                   `json:"paymentToken,omitempty" jsonapi:"attr,paymentToken,omitempty" valid:"optional"`           // Example: "f106ac99-213c-4cf7-8c1b-1e841516026b"
	PriceIncludingVAT float32                  `json:"priceIncludingVAT,omitempty" jsonapi:"attr,priceIncludingVAT,omitempty" valid:"optional"` // Example: "69.34"
	PriceWithoutVAT   float32                  `json:"priceWithoutVAT,omitempty" jsonapi:"attr,priceWithoutVAT,omitempty" valid:"optional"`     // Example: "58.27"
	PumpID            string                   `json:"pumpId,omitempty" jsonapi:"attr,pumpId,omitempty" valid:"optional,uuid"`                  // Example: "460ffaad-a3c1-4199-b69e-63949ccda82f"
	Vin               string                   `json:"vin,omitempty" jsonapi:"attr,vin,omitempty" valid:"optional"`                             // Example: "1B3EL46R36N102271"
}

// ProcessPaymentCreatedVAT ...
type ProcessPaymentCreatedVAT struct {
	Amount float32 `json:"amount,omitempty" jsonapi:"attr,amount,omitempty" valid:"optional"` // Example: "11.07"
	Rate   float32 `json:"rate,omitempty" jsonapi:"attr,rate,omitempty" valid:"optional"`     // Example: "0.19"
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
	Request           *http.Request      `valid:"-"`
	Content           TransactionRequest `valid:"-"`
	ParamGasStationID string             `valid:"required,uuid"`
}

/*
ApproachingAtTheForecourtResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type ApproachingAtTheForecourtResponseWriter interface {
	http.ResponseWriter
	Created(ApproachingResponse)
	BadRequest(error)
	NotFound(error)
}
type approachingAtTheForecourtResponseWriter struct {
	http.ResponseWriter
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *approachingAtTheForecourtResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *approachingAtTheForecourtResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// Created responds with jsonapi marshaled data (HTTP code 201)
func (w *approachingAtTheForecourtResponseWriter) Created(data ApproachingResponse) {
	runtime.Marshal(w, data, 201)
}

// ApproachingAtTheForecourtRequest ...
type ApproachingAtTheForecourtRequest struct {
	Request           *http.Request      `valid:"-"`
	Content           ApproachingRequest `valid:"-"`
	ParamGasStationID string             `valid:"required,uuid"`
}

/*
GetPumpResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetPumpResponseWriter interface {
	http.ResponseWriter
	OK(PumpResponse)
	NotFound(error)
}
type getPumpResponseWriter struct {
	http.ResponseWriter
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *getPumpResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *getPumpResponseWriter) OK(data PumpResponse) {
	runtime.Marshal(w, data, 200)
}

/*
GetPumpRequest is a standard http.Request extended with the
un-marshaled content object
*/
type GetPumpRequest struct {
	Request           *http.Request `valid:"-"`
	ParamGasStationID string        `valid:"required,uuid"`
	ParamPumpID       string        `valid:"required,uuid"`
}

/*
WaitOnPumpStatusChangeResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type WaitOnPumpStatusChangeResponseWriter interface {
	http.ResponseWriter
	OK(PumpResponse)
	BadRequest(error)
	NotFound(error)
	RequestTimeout(error)
}
type waitOnPumpStatusChangeResponseWriter struct {
	http.ResponseWriter
}

// RequestTimeout responds with jsonapi error (HTTP code 408)
func (w *waitOnPumpStatusChangeResponseWriter) RequestTimeout(err error) {
	runtime.WriteError(w, 408, err)
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *waitOnPumpStatusChangeResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *waitOnPumpStatusChangeResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *waitOnPumpStatusChangeResponseWriter) OK(data PumpResponse) {
	runtime.Marshal(w, data, 200)
}

/*
WaitOnPumpStatusChangeRequest is a standard http.Request extended with the
un-marshaled content object
*/
type WaitOnPumpStatusChangeRequest struct {
	Request           *http.Request `valid:"-"`
	ParamGasStationID string        `valid:"required,uuid"`
	ParamPumpID       string        `valid:"required,uuid"`
	ParamUpdate       string        `valid:"required,in(longPolling)"`
	ParamLastStatus   PumpStatus    `valid:"optional"`
	ParamTimeout      int64         `valid:"optional"`
}

// Service interface for all handlers
type Service interface {
	/*
	   ProcessPayment Process payment

	   Process payment and notify user if transaction is finished successfully. You can optionally provide `priceIncludingVAT`and `currency` in the request body to check if the price the user has seen is still correct.
	*/
	ProcessPayment(context.Context, ProcessPaymentResponseWriter, *ProcessPaymentRequest) error
	/*
	   ApproachingAtTheForecourt Gather information when approaching at the forecourt


	   This request will:
	   * Return a list of available paymentMethodIds
	   * Return up-to-date price information (price structure) at the gas station
	   * Return a list of pumps available at the gas station together with the current status (free, inUse, readyToPay, outOfOrder)
	   * Create payment tokens for all paymentMethods of the user and pre-authorise the calculated maximum amount of money (background task)
	*/
	ApproachingAtTheForecourt(context.Context, ApproachingAtTheForecourtResponseWriter, *ApproachingAtTheForecourtRequest) error
	/*
	   GetPump Return current pump information

	   Returns the current pump status (free, inUse, readyToPay, outOfOrder) and identifier. If the status is readyToPay, the result also contains fuelType, productName, fuelAmount, VAT (amount & rate), priceWithoutVAT, priceIncludingVAT, currency.
	*/
	GetPump(context.Context, GetPumpResponseWriter, *GetPumpRequest) error
	/*
	   WaitOnPumpStatusChange Wait for a status change on a given pump

	   Uses **long polling** to wait for a status change on a given pump. Returns as soon as the status has changed or after the number of seconds provided by the optional `timeout` query parameter (default timeout is 30 seconds). In case of timeout (408 status code) you're safe to start the request again. Instantaneously returns if `lastStatus` was given and already changed between request. If successful, it returns the same structure as the normal status call
	*/
	WaitOnPumpStatusChange(context.Context, WaitOnPumpStatusChangeResponseWriter, *WaitOnPumpStatusChangeRequest) error
}

/*
Router implements: PACE Fueling API

Fueling API
*/
func Router(service Service) *mux.Router {
	router := mux.NewRouter()
	// Subrouter s1 - Path: /fueling
	s1 := router.PathPrefix("/fueling").Subrouter()
	s1.Methods("GET").Path("/beta/gas-stations/{gasStationId}/pumps/{pumpId}/wait-for-status-change").Handler(WaitOnPumpStatusChangeHandler(service)).Name("WaitOnPumpStatusChange")
	s1.Methods("GET").Path("/beta/gas-stations/{gasStationId}/pumps/{pumpId}").Handler(GetPumpHandler(service)).Name("GetPump")
	s1.Methods("POST").Path("/beta/gas-station/{gasStationId}/payment").Handler(ProcessPaymentHandler(service)).Name("ProcessPayment")
	s1.Methods("POST").Path("/beta/gas-stations/{gasStationId}/approaching").Handler(ApproachingAtTheForecourtHandler(service)).Name("ApproachingAtTheForecourt")
	return router
}
