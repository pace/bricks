package fueling

import (
	"context"
	"encoding/json"
	"errors"
	mux "github.com/gorilla/mux"
	opentracing "github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
	runtime "lab.jamit.de/pace/go-microservice/http/jsonapi/runtime"
	log "lab.jamit.de/pace/go-microservice/maintenance/log"
	jsonapimetrics "lab.jamit.de/pace/go-microservice/maintenance/metrics/jsonapi"
	"net/http"
)

// FuelPrice ...
type FuelPrice struct {
	ID             string          `jsonapi:"primary,fuelPrice,omitempty" valid:"optional"` // Fuel Price ID
	Currency       *Currency       `json:"currency,omitempty" jsonapi:"attr,currency,omitempty" valid:"optional"`
	FuelAmountUnit *FuelAmountUnit `json:"fuelAmountUnit,omitempty" jsonapi:"attr,fuelAmountUnit,omitempty" valid:"optional"`
	FuelType       string          `json:"fuelType,omitempty" jsonapi:"attr,fuelType,omitempty" valid:"optional"`         // Example: "ron95_e10"
	PricePerUnit   float32         `json:"pricePerUnit,omitempty" jsonapi:"attr,pricePerUnit,omitempty" valid:"optional"` // Example: "1.379"
	ProductName    string          `json:"productName,omitempty" jsonapi:"attr,productName,omitempty" valid:"optional"`   // Example: "Super E10"
}

// FuelPriceResponse ...
type FuelPriceResponse *FuelPrice

// GasStationResponseItem ...
type GasStationResponseItem struct {
	ID             string                            `jsonapi:"primary,gasStation,omitempty" valid:"uuid,optional"` // Gas Station ID
	Address        *GasStationResponseAddress        `json:"address,omitempty" jsonapi:"attr,address,omitempty" valid:"optional"`
	Amenities      []string                          `json:"amenities,omitempty" jsonapi:"attr,amenities,omitempty" valid:"optional"` // Example: "[restaurant]"
	Latitude       float32                           `json:"latitude,omitempty" jsonapi:"attr,latitude,omitempty" valid:"optional"`   // Example: "49.013"
	Longitude      float32                           `json:"longitude,omitempty" jsonapi:"attr,longitude,omitempty" valid:"optional"` // Example: "8.425"
	OpeningHours   []*GasStationResponseOpeningHours `json:"openingHours,omitempty" jsonapi:"attr,openingHours,omitempty" valid:"optional"`
	StationName    string                            `json:"stationName,omitempty" jsonapi:"attr,stationName,omitempty" valid:"optional"` // Example: "PACE Station"
	FuelPrices     []*FuelPrice                      `json:"fuelPrices,omitempty" jsonapi:"attr,fuelPrices,omitempty" valid:"optional"`
	PaymentMethods []*PaymentMethod                  `json:"paymentMethods,omitempty" jsonapi:"attr,paymentMethods,omitempty" valid:"optional"`
	Pumps          []*Pump                           `json:"pumps,omitempty" jsonapi:"attr,pumps,omitempty" valid:"optional"`
}

// GasStationResponseAddress ...
type GasStationResponseAddress struct {
	City        string `json:"city,omitempty" jsonapi:"city,omitempty" valid:"optional"`               // Example: "Karlsruhe"
	CountryCode string `json:"countryCode,omitempty" jsonapi:"countryCode,omitempty" valid:"optional"` // Country code in as specified in ISO 3166-1.
	HouseNo     string `json:"houseNo,omitempty" jsonapi:"houseNo,omitempty" valid:"optional"`         // Example: "18"
	PostalCode  string `json:"postalCode,omitempty" jsonapi:"postalCode,omitempty" valid:"optional"`   // Example: "76131"
	Street      string `json:"street,omitempty" jsonapi:"street,omitempty" valid:"optional"`           // Example: "Haid-und-Neu-Str."
}

// GasStationResponseOpeningHours ...
type GasStationResponseOpeningHours struct {
	OpenFromTo []string `json:"openFromTo,omitempty" jsonapi:"openFromTo,omitempty" valid:"optional"` // Example: "[07:30 20:30]"
	Weekdays   []string `json:"weekdays,omitempty" jsonapi:"weekdays,omitempty" valid:"optional"`     // Example: "[Montag Dienstag]"
}

// GasStationResponse ...
type GasStationResponse []*GasStationResponseItem

// PaymentMethod ...
type PaymentMethod struct {
	ID   string `jsonapi:"primary,paymentMethod,omitempty" valid:"optional"`           // Payment Method ID
	Kind string `json:"kind,omitempty" jsonapi:"attr,kind,omitempty" valid:"optional"` // Example: "sepa"
}

// PaymentMethodResponse ...
type PaymentMethodResponse *PaymentMethod

// Pump ...
type Pump struct {
	ID         string      `jsonapi:"primary,pump,omitempty" valid:"uuid,optional"`                           // Pump ID
	Identifier string      `json:"identifier,omitempty" jsonapi:"attr,identifier,omitempty" valid:"optional"` // Pump identifier
	Status     *PumpStatus `json:"status,omitempty" jsonapi:"attr,status,omitempty" valid:"optional"`
}

// PumpReadyForPaymentResponse ...
type PumpReadyForPaymentResponse json.RawMessage

// PumpResponse ...
type PumpResponse *Pump

// PumpStatus Current pump status
type PumpStatus string

// Currency ...
type Currency string

// FuelAmountUnit ...
type FuelAmountUnit string

/*
ApproachingAtTheForecourtHandler handles request/response marshaling and validation for
 Get /beta/gas-station/{fuelingAppId}/approaching
*/
func ApproachingAtTheForecourtHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer func() {
			if rp := recover(); rp != nil {
				log.Ctx(ctx).Error().Str("handler", "ApproachingAtTheForecourtHandler").Msgf("Panic: %v", rp)
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
		handlerSpan = opentracing.StartSpan("ApproachingAtTheForecourtHandler", opentracing.ChildOf(wireContext))
		handlerSpan.LogFields(olog.String("req_id", log.RequestID(r)))
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		ctx = opentracing.ContextWithSpan(r.Context(), handlerSpan)
		writer := approachingAtTheForecourtResponseWriter{
			ResponseWriter: jsonapimetrics.NewMetric("fueling", "/beta/gas-station/{fuelingAppId}/approaching", w, r),
		}
		request := ApproachingAtTheForecourtRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamFuelingAppID,
			Location: runtime.ScanInPath,
			Input:    vars["fuelingAppId"],
			Name:     "fuelingAppId",
		}, &runtime.ScanParameter{
			Data:     &request.ParamExpectedAmountInLiters,
			Location: runtime.ScanInQuery,
			Input:    vars["expectedAmountInLiters"],
			Name:     "expectedAmountInLiters",
		}, &runtime.ScanParameter{
			Data:     &request.ParamCarFuelType,
			Location: runtime.ScanInQuery,
			Input:    vars["carFuelType"],
			Name:     "carFuelType",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Invoke service that implements the business logic
		err = service.ApproachingAtTheForecourt(ctx, &writer, &request)
		if err != nil {
			runtime.WriteError(w, http.StatusInternalServerError, err)
		}
	})
}

/*
GetPumpHandler handles request/response marshaling and validation for
 Get /beta/gas-station/{fuelingAppId}/pumps/{pumpId}
*/
func GetPumpHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer func() {
			if rp := recover(); rp != nil {
				log.Ctx(ctx).Error().Str("handler", "GetPumpHandler").Msgf("Panic: %v", rp)
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
		handlerSpan = opentracing.StartSpan("GetPumpHandler", opentracing.ChildOf(wireContext))
		handlerSpan.LogFields(olog.String("req_id", log.RequestID(r)))
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		ctx = opentracing.ContextWithSpan(r.Context(), handlerSpan)
		writer := getPumpResponseWriter{
			ResponseWriter: jsonapimetrics.NewMetric("fueling", "/beta/gas-station/{fuelingAppId}/pumps/{pumpId}", w, r),
		}
		request := GetPumpRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamFuelingAppID,
			Location: runtime.ScanInPath,
			Input:    vars["fuelingAppId"],
			Name:     "fuelingAppId",
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
		err = service.GetPump(ctx, &writer, &request)
		if err != nil {
			runtime.WriteError(w, http.StatusInternalServerError, err)
		}
	})
}

/*
WaitOnPumpStatusChangeHandler handles request/response marshaling and validation for
 Get /beta/gas-station/{fuelingAppId}/pumps/{pumpId}/wait-for-status-change
*/
func WaitOnPumpStatusChangeHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer func() {
			if rp := recover(); rp != nil {
				log.Ctx(ctx).Error().Str("handler", "WaitOnPumpStatusChangeHandler").Msgf("Panic: %v", rp)
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
		handlerSpan = opentracing.StartSpan("WaitOnPumpStatusChangeHandler", opentracing.ChildOf(wireContext))
		handlerSpan.LogFields(olog.String("req_id", log.RequestID(r)))
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		ctx = opentracing.ContextWithSpan(r.Context(), handlerSpan)
		writer := waitOnPumpStatusChangeResponseWriter{
			ResponseWriter: jsonapimetrics.NewMetric("fueling", "/beta/gas-station/{fuelingAppId}/pumps/{pumpId}/wait-for-status-change", w, r),
		}
		request := WaitOnPumpStatusChangeRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamFuelingAppID,
			Location: runtime.ScanInPath,
			Input:    vars["fuelingAppId"],
			Name:     "fuelingAppId",
		}, &runtime.ScanParameter{
			Data:     &request.ParamPumpID,
			Location: runtime.ScanInPath,
			Input:    vars["pumpId"],
			Name:     "pumpId",
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
		err = service.WaitOnPumpStatusChange(ctx, &writer, &request)
		if err != nil {
			runtime.WriteError(w, http.StatusInternalServerError, err)
		}
	})
}

/*
ApproachingAtTheForecourtResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type ApproachingAtTheForecourtResponseWriter interface {
	http.ResponseWriter
	OK(GasStationResponse)
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

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *approachingAtTheForecourtResponseWriter) OK(data GasStationResponse) {
	runtime.Marshal(w, data, 200)
}

/*
ApproachingAtTheForecourtResponseWriter is a standard http.Request extended with the
un-marshaled content object
*/
type ApproachingAtTheForecourtRequest struct {
	Request                     *http.Request `valid:"-"`
	ParamFuelingAppID           string        `valid:"required,uuid"`
	ParamExpectedAmountInLiters float32       `valid:"required"`
	ParamCarFuelType            string        `valid:"required,in(e85|ron91|ron95_e5|ron95_e10|ron98|ron98_e5|ron100|diesel|diesel_gtl|diesel_b7|lpg|cng|h2|Truck Diesel|AdBlue)"`
}

// GetPumpOK ...
type GetPumpOK json.RawMessage

/*
GetPumpResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetPumpResponseWriter interface {
	http.ResponseWriter
	OK(*GetPumpOK)
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
func (w *getPumpResponseWriter) OK(data *GetPumpOK) {
	runtime.Marshal(w, data, 200)
}

/*
GetPumpResponseWriter is a standard http.Request extended with the
un-marshaled content object
*/
type GetPumpRequest struct {
	Request           *http.Request `valid:"-"`
	ParamFuelingAppID string        `valid:"required,uuid"`
	ParamPumpID       string        `valid:"required,uuid"`
}

// WaitOnPumpStatusChangeOK ...
type WaitOnPumpStatusChangeOK json.RawMessage

/*
WaitOnPumpStatusChangeResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type WaitOnPumpStatusChangeResponseWriter interface {
	http.ResponseWriter
	OK(*WaitOnPumpStatusChangeOK)
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
func (w *waitOnPumpStatusChangeResponseWriter) OK(data *WaitOnPumpStatusChangeOK) {
	runtime.Marshal(w, data, 200)
}

/*
WaitOnPumpStatusChangeResponseWriter is a standard http.Request extended with the
un-marshaled content object
*/
type WaitOnPumpStatusChangeRequest struct {
	Request           *http.Request `valid:"-"`
	ParamFuelingAppID string        `valid:"required,uuid"`
	ParamPumpID       string        `valid:"required,uuid"`
	ParamLastStatus   PumpStatus    `valid:"optional"`
	ParamTimeout      int64         `valid:"optional"`
}
type Service interface {
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

	   Returns the current pump status (free, inUse, readyToPay, outOfOrder) and identifier. If the status is readyToPay, the result also contains fuelType, productName, fuelAmount, fuelAmountUnit, VAT (amount & rate), priceWithoutVAT, priceIncludingVAT, currency.
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
	// Subrouter s1 - https://api.pace.cloud/fueling
	s1 := router.PathPrefix("/fueling").Subrouter()
	s1.Methods("GET").Path("/beta/gas-station/{fuelingAppId}/pumps/{pumpId}/wait-for-status-change").Handler(WaitOnPumpStatusChangeHandler(service)).Name("WaitOnPumpStatusChange")
	s1.Methods("GET").Path("/beta/gas-station/{fuelingAppId}/pumps/{pumpId}").Handler(GetPumpHandler(service)).Name("GetPump")
	s1.Methods("GET").Path("/beta/gas-station/{fuelingAppId}/approaching").Handler(ApproachingAtTheForecourtHandler(service)).Name("ApproachingAtTheForecourt")
	return router
}
