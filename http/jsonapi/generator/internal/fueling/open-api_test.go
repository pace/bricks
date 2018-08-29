package fueling

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	mux "github.com/gorilla/mux"
	runtime "lab.jamit.de/pace/web/libs/go-microservice/http/jsonapi/runtime"
	"net/http"
	"runtime/debug"
)

// FuelPrice ...
type FuelPrice struct {
	ID             string          `jsonapi:"primary,fuelPrice,omitempty" valid:"optional"`                                   // Fuel Price ID
	Currency       *Currency       `json:"currency,omitempty" jsonapi:"attr,currency,omitempty" valid:"optional"`             // Example: "EUR"
	FuelAmountUnit *FuelAmountUnit `json:"fuelAmountUnit,omitempty" jsonapi:"attr,fuelAmountUnit,omitempty" valid:"optional"` // Example: "Ltr"
	FuelType       string          `json:"fuelType,omitempty" jsonapi:"attr,fuelType,omitempty" valid:"optional"`             // Example: "ron95_e10"
	PricePerUnit   float32         `json:"pricePerUnit,omitempty" jsonapi:"attr,pricePerUnit,omitempty" valid:"optional"`     // Example: "1.379"
	ProductName    string          `json:"productName,omitempty" jsonapi:"attr,productName,omitempty" valid:"optional"`       // Example: "Super E10"
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
	PaymentMethods []string                          `json:"paymentMethods,omitempty" jsonapi:"attr,paymentMethods,omitempty" valid:"optional"` // Example: "[sepaDirectDebit]"
	StationName    string                            `json:"stationName,omitempty" jsonapi:"attr,stationName,omitempty" valid:"optional"`       // Example: "PACE Station"
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
	Status     *PumpStatus `json:"status,omitempty" jsonapi:"attr,status,omitempty" valid:"optional"`         // Current pump status
}

// PumpReadyForPaymentResponse ...
type PumpReadyForPaymentResponse json.RawMessage

// PumpResponse ...
type PumpResponse *Pump

// PumpStatus Current pump status
type PumpStatus string

// TransactionRequest ...
type TransactionRequest struct {
	ID              string `jsonapi:"primary,transaction,omitempty" valid:"uuid,optional"`                              // Transaction ID
	MileageInMeters int64  `json:"mileageInMeters,omitempty" jsonapi:"attr,mileageInMeters,omitempty" valid:"required"` // Example: "66435"
	PaymentMethodID string `json:"paymentMethodId,omitempty" jsonapi:"attr,paymentMethodId,omitempty" valid:"required"` // Example: "f106ac99-213c-4cf7-8c1b-1e841516026b"
	Vin             string `json:"vin,omitempty" jsonapi:"attr,vin,omitempty" valid:"required"`                         // Example: "1B3EL46R36N102271"
}

// TransactionWithPriceCheckRequest ...
type TransactionWithPriceCheckRequest json.RawMessage

// Currency ...
type Currency string

// FuelAmountUnit ...
type FuelAmountUnit string

/*
GetGasStationFuelingAppIDApproachingHandler handles request/response marshaling and validation for
 Get /gas-station/{fuelingAppId}/approaching
*/
func GetGasStationFuelingAppIDApproachingHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Panic %s: %v\n", "GetGasStationFuelingAppIDApproachingHandler", r)
				debug.PrintStack()
				runtime.WriteError(w, http.StatusInternalServerError, errors.New("Error"))
			}
		}()
		writer := getGasStationFuelingAppIDApproachingResponseWriter{
			ResponseWriter: w,
		}
		request := GetGasStationFuelingAppIDApproachingRequest{
			Request: r,
		}
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
		err := service.GetGasStationFuelingAppIDApproaching(r.Context(), &writer, &request)
		if err != nil {
			runtime.WriteError(w, http.StatusInternalServerError, err)
		}
	})
}

/*
GetGasStationFuelingAppIDPumpsPumpIDHandler handles request/response marshaling and validation for
 Get /gas-station/{fuelingAppId}/pumps/{pumpId}
*/
func GetGasStationFuelingAppIDPumpsPumpIDHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Panic %s: %v\n", "GetGasStationFuelingAppIDPumpsPumpIDHandler", r)
				debug.PrintStack()
				runtime.WriteError(w, http.StatusInternalServerError, errors.New("Error"))
			}
		}()
		writer := getGasStationFuelingAppIDPumpsPumpIDResponseWriter{
			ResponseWriter: w,
		}
		request := GetGasStationFuelingAppIDPumpsPumpIDRequest{
			Request: r,
		}
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
		err := service.GetGasStationFuelingAppIDPumpsPumpID(r.Context(), &writer, &request)
		if err != nil {
			runtime.WriteError(w, http.StatusInternalServerError, err)
		}
	})
}

/*
PostGasStationFuelingAppIDPumpsPumpIDPayHandler handles request/response marshaling and validation for
 Post /gas-station/{fuelingAppId}/pumps/{pumpId}/pay
*/
func PostGasStationFuelingAppIDPumpsPumpIDPayHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Panic %s: %v\n", "PostGasStationFuelingAppIDPumpsPumpIDPayHandler", r)
				debug.PrintStack()
				runtime.WriteError(w, http.StatusInternalServerError, errors.New("Error"))
			}
		}()
		writer := postGasStationFuelingAppIDPumpsPumpIDPayResponseWriter{
			ResponseWriter: w,
		}
		request := PostGasStationFuelingAppIDPumpsPumpIDPayRequest{
			Request: r,
		}
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
		if runtime.Unmarshal(w, r, &request.Content) {
			err := service.PostGasStationFuelingAppIDPumpsPumpIDPay(r.Context(), &writer, &request)
			if err != nil {
				runtime.WriteError(w, http.StatusInternalServerError, err)
			}
		}
	})
}

/*
GetGasStationFuelingAppIDPumpsPumpIDWaitForStatusChangeHandler handles request/response marshaling and validation for
 Get /gas-station/{fuelingAppId}/pumps/{pumpId}/waitForStatusChange
*/
func GetGasStationFuelingAppIDPumpsPumpIDWaitForStatusChangeHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Panic %s: %v\n", "GetGasStationFuelingAppIDPumpsPumpIDWaitForStatusChangeHandler", r)
				debug.PrintStack()
				runtime.WriteError(w, http.StatusInternalServerError, errors.New("Error"))
			}
		}()
		writer := getGasStationFuelingAppIDPumpsPumpIDWaitForStatusChangeResponseWriter{
			ResponseWriter: w,
		}
		request := GetGasStationFuelingAppIDPumpsPumpIDWaitForStatusChangeRequest{
			Request: r,
		}
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
		err := service.GetGasStationFuelingAppIDPumpsPumpIDWaitForStatusChange(r.Context(), &writer, &request)
		if err != nil {
			runtime.WriteError(w, http.StatusInternalServerError, err)
		}
	})
}

/*
GetGasStationFuelingAppIDApproachingResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetGasStationFuelingAppIDApproachingResponseWriter interface {
	http.ResponseWriter
	OK(GasStationResponse)
	NotFound(error)
}
type getGasStationFuelingAppIDApproachingResponseWriter struct {
	http.ResponseWriter
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *getGasStationFuelingAppIDApproachingResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *getGasStationFuelingAppIDApproachingResponseWriter) OK(data GasStationResponse) {
	runtime.Marshal(w, data, 200)
}

/*
GetGasStationFuelingAppIDApproachingResponseWriter is a standard http.Request extended with the
un-marshaled content object
*/
type GetGasStationFuelingAppIDApproachingRequest struct {
	Request                     *http.Request `valid:"-"`
	ParamFuelingAppID           string        `valid:"required,uuid"`
	ParamExpectedAmountInLiters float32       `valid:"required"`
	ParamCarFuelType            string        `valid:"required,in(e85|ron91|ron95_e5|ron95_e10|ron98|ron98_e5|ron100|diesel|diesel_gtl|diesel_b7|lpg|cng|h2|Truck Diesel|AdBlue)"`
}

// GetGasStationFuelingAppIDPumpsPumpIDOK ...
type GetGasStationFuelingAppIDPumpsPumpIDOK json.RawMessage

/*
GetGasStationFuelingAppIDPumpsPumpIDResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetGasStationFuelingAppIDPumpsPumpIDResponseWriter interface {
	http.ResponseWriter
	OK(*GetGasStationFuelingAppIDPumpsPumpIDOK)
	NotFound(error)
}
type getGasStationFuelingAppIDPumpsPumpIDResponseWriter struct {
	http.ResponseWriter
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *getGasStationFuelingAppIDPumpsPumpIDResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *getGasStationFuelingAppIDPumpsPumpIDResponseWriter) OK(data *GetGasStationFuelingAppIDPumpsPumpIDOK) {
	runtime.Marshal(w, data, 200)
}

/*
GetGasStationFuelingAppIDPumpsPumpIDResponseWriter is a standard http.Request extended with the
un-marshaled content object
*/
type GetGasStationFuelingAppIDPumpsPumpIDRequest struct {
	Request           *http.Request `valid:"-"`
	ParamFuelingAppID string        `valid:"required,uuid"`
	ParamPumpID       string        `valid:"required,uuid"`
}

// PostGasStationFuelingAppIDPumpsPumpIDPayCreated ...
type PostGasStationFuelingAppIDPumpsPumpIDPayCreated struct {
	ID                string                                              `jsonapi:"primary,transaction,omitempty" valid:"optional"` // Transaction ID
	VAT               *PostGasStationFuelingAppIDPumpsPumpIDPayCreatedVAT `json:"VAT,omitempty" jsonapi:"attr,VAT,omitempty" valid:"optional"`
	Currency          *Currency                                           `json:"currency,omitempty" jsonapi:"attr,currency,omitempty" valid:"optional"`                   // Example: "EUR"
	FuelingAppID      string                                              `json:"fuelingAppId,omitempty" jsonapi:"attr,fuelingAppId,omitempty" valid:"optional"`           // Example: "c30bce97-b732-4390-af38-1ac6b017aa4c"
	MileageInMeters   int64                                               `json:"mileageInMeters,omitempty" jsonapi:"attr,mileageInMeters,omitempty" valid:"optional"`     // Example: "66435"
	PaymentMethodID   string                                              `json:"paymentMethodId,omitempty" jsonapi:"attr,paymentMethodId,omitempty" valid:"optional"`     // Example: "f106ac99-213c-4cf7-8c1b-1e841516026b"
	PriceIncludingVAT float32                                             `json:"priceIncludingVAT,omitempty" jsonapi:"attr,priceIncludingVAT,omitempty" valid:"optional"` // Example: "69.34"
	PriceWithoutVAT   float32                                             `json:"priceWithoutVAT,omitempty" jsonapi:"attr,priceWithoutVAT,omitempty" valid:"optional"`     // Example: "58.27"
	PumpID            string                                              `json:"pumpId,omitempty" jsonapi:"attr,pumpId,omitempty" valid:"optional"`                       // Example: "460ffaad-a3c1-4199-b69e-63949ccda82f"
	Vin               string                                              `json:"vin,omitempty" jsonapi:"attr,vin,omitempty" valid:"optional"`                             // Example: "1B3EL46R36N102271"
}

// PostGasStationFuelingAppIDPumpsPumpIDPayCreatedVAT ...
type PostGasStationFuelingAppIDPumpsPumpIDPayCreatedVAT struct {
	Amount float32 `json:"amount,omitempty" jsonapi:"amount,omitempty" valid:"optional"` // Example: "11.07"
	Rate   float32 `json:"rate,omitempty" jsonapi:"rate,omitempty" valid:"optional"`     // Example: "0.19"
}

/*
PostGasStationFuelingAppIDPumpsPumpIDPayResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type PostGasStationFuelingAppIDPumpsPumpIDPayResponseWriter interface {
	http.ResponseWriter
	Created(*PostGasStationFuelingAppIDPumpsPumpIDPayCreated)
	BadRequest(error)
	NotFound(error)
	Conflict(error)
}
type postGasStationFuelingAppIDPumpsPumpIDPayResponseWriter struct {
	http.ResponseWriter
}

// Conflict responds with jsonapi error (HTTP code 409)
func (w *postGasStationFuelingAppIDPumpsPumpIDPayResponseWriter) Conflict(err error) {
	runtime.WriteError(w, 409, err)
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *postGasStationFuelingAppIDPumpsPumpIDPayResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *postGasStationFuelingAppIDPumpsPumpIDPayResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// Created responds with jsonapi marshaled data (HTTP code 201)
func (w *postGasStationFuelingAppIDPumpsPumpIDPayResponseWriter) Created(data *PostGasStationFuelingAppIDPumpsPumpIDPayCreated) {
	runtime.Marshal(w, data, 201)
}

// PostGasStationFuelingAppIDPumpsPumpIDPayContent ...
type PostGasStationFuelingAppIDPumpsPumpIDPayContent json.RawMessage

// PostGasStationFuelingAppIDPumpsPumpIDPayRequest ...
type PostGasStationFuelingAppIDPumpsPumpIDPayRequest struct {
	Request           *http.Request                                    `valid:"-"`
	Content           *PostGasStationFuelingAppIDPumpsPumpIDPayContent `valid:"-"`
	ParamFuelingAppID string                                           `valid:"required,uuid"`
	ParamPumpID       string                                           `valid:"required,uuid"`
}

// GetGasStationFuelingAppIDPumpsPumpIDWaitForStatusChangeOK ...
type GetGasStationFuelingAppIDPumpsPumpIDWaitForStatusChangeOK json.RawMessage

/*
GetGasStationFuelingAppIDPumpsPumpIDWaitForStatusChangeResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetGasStationFuelingAppIDPumpsPumpIDWaitForStatusChangeResponseWriter interface {
	http.ResponseWriter
	OK(*GetGasStationFuelingAppIDPumpsPumpIDWaitForStatusChangeOK)
	NotFound(error)
	RequestTimeout(error)
}
type getGasStationFuelingAppIDPumpsPumpIDWaitForStatusChangeResponseWriter struct {
	http.ResponseWriter
}

// RequestTimeout responds with jsonapi error (HTTP code 408)
func (w *getGasStationFuelingAppIDPumpsPumpIDWaitForStatusChangeResponseWriter) RequestTimeout(err error) {
	runtime.WriteError(w, 408, err)
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *getGasStationFuelingAppIDPumpsPumpIDWaitForStatusChangeResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *getGasStationFuelingAppIDPumpsPumpIDWaitForStatusChangeResponseWriter) OK(data *GetGasStationFuelingAppIDPumpsPumpIDWaitForStatusChangeOK) {
	runtime.Marshal(w, data, 200)
}

/*
GetGasStationFuelingAppIDPumpsPumpIDWaitForStatusChangeResponseWriter is a standard http.Request extended with the
un-marshaled content object
*/
type GetGasStationFuelingAppIDPumpsPumpIDWaitForStatusChangeRequest struct {
	Request           *http.Request `valid:"-"`
	ParamFuelingAppID string        `valid:"required,uuid"`
	ParamPumpID       string        `valid:"required,uuid"`
	ParamLastStatus   string        `valid:"optional,in(free|inUse|readyToPay|outOfOrder)"`
	ParamTimeout      int64         `valid:"optional"`
}
type Service interface {
	/*
	   GetGasStationFuelingAppIDApproaching Gather information when approaching at the forecourt


	   This request will:
	   * Return a list of available paymentMethodIds
	   * Return up-to-date price information (price structure) at the gas station
	   * Return a list of pumps available at the gas station together with the current status (free, inUse, readyToPay, outOfOrder)
	   * Create payment tokens for all paymentMethods of the user and pre-authorise the calculated maximum amount of money (background task)
	*/
	GetGasStationFuelingAppIDApproaching(context.Context, GetGasStationFuelingAppIDApproachingResponseWriter, *GetGasStationFuelingAppIDApproachingRequest) error
	/*
	   GetGasStationFuelingAppIDPumpsPumpID Return current pump information

	   Returns the current pump status (free, inUse, readyToPay, outOfOrder) and identifier. If the status is readyToPay, the result also contains fuelType, productName, fuelAmount, fuelAmountUnit, VAT (amount & rate), priceWithoutVAT, priceIncludingVAT, currency.
	*/
	GetGasStationFuelingAppIDPumpsPumpID(context.Context, GetGasStationFuelingAppIDPumpsPumpIDResponseWriter, *GetGasStationFuelingAppIDPumpsPumpIDRequest) error
	/*
	   PostGasStationFuelingAppIDPumpsPumpIDPay Process payment

	   Process payment and notify user if transaction is finished successfully. You can optionally provide `priceIncludingVAT`and `currency` in the request body to check if the price the user has seen is still correct.
	*/
	PostGasStationFuelingAppIDPumpsPumpIDPay(context.Context, PostGasStationFuelingAppIDPumpsPumpIDPayResponseWriter, *PostGasStationFuelingAppIDPumpsPumpIDPayRequest) error
	/*
	   GetGasStationFuelingAppIDPumpsPumpIDWaitForStatusChange Wait for a status change on a given pump

	   Uses **long polling** to wait for a status change on a given pump. Returns as soon as the status has changed or after the number of seconds provided by the optional `timeout` query parameter (default timeout is 30 seconds). In case of timeout (408 status code) you're safe to start the request again. Instantaneously returns if `lastStatus` was given and already changed between request. If successful, it returns the same structure as the normal status call
	*/
	GetGasStationFuelingAppIDPumpsPumpIDWaitForStatusChange(context.Context, GetGasStationFuelingAppIDPumpsPumpIDWaitForStatusChangeResponseWriter, *GetGasStationFuelingAppIDPumpsPumpIDWaitForStatusChangeRequest) error
}

/*
Router implements: PACE Fueling API

Fueling API
*/
func Router(service Service) *mux.Router {
	router := mux.NewRouter()
	// Subrouter s1 - https://api.pace.cloud/fueling/beta
	s1 := router.PathPrefix("/fueling/beta").Subrouter()
	s1.Methods("POST").Path("/gas-station/{fuelingAppId}/pumps/{pumpId}/pay").Handler(PostGasStationFuelingAppIDPumpsPumpIDPayHandler(service)).Name("PostGasStationFuelingAppIDPumpsPumpIDPay")
	s1.Methods("GET").Path("/gas-station/{fuelingAppId}/pumps/{pumpId}/waitForStatusChange").Handler(GetGasStationFuelingAppIDPumpsPumpIDWaitForStatusChangeHandler(service)).Name("GetGasStationFuelingAppIDPumpsPumpIDWaitForStatusChange")
	s1.Methods("GET").Path("/gas-station/{fuelingAppId}/pumps/{pumpId}").Handler(GetGasStationFuelingAppIDPumpsPumpIDHandler(service)).Name("GetGasStationFuelingAppIDPumpsPumpID")
	s1.Methods("GET").Path("/gas-station/{fuelingAppId}/approaching").Handler(GetGasStationFuelingAppIDApproachingHandler(service)).Name("GetGasStationFuelingAppIDApproaching")
	return router
}
