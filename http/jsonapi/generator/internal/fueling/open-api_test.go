package fueling

import (
	"context"
	mux "github.com/gorilla/mux"
	runtime "lab.jamit.de/pace/web/libs/go-microservice/http/jsonapi/runtime"
	"net/http"
)

// FuelPrice ...
type FuelPrice struct {
	Attributes *FuelPriceAttributes `jsonapi:"attributes,omitempty" valid:"optional"`
	Id         string               `jsonapi:"id,omitempty" valid:"optional"`                 // Fuel Price ID
	Type       string               `jsonapi:"type,omitempty" valid:"optional,in(fuelPrice)"` // Fuel price
}

// FuelPriceAttributes ...
type FuelPriceAttributes struct {
	Currency       string  `jsonapi:"currency,omitempty" valid:"optional,in(EUR)"`                                                                                                           // Example: "EUR"
	FuelAmountUnit string  `jsonapi:"fuelAmountUnit,omitempty" valid:"optional,in(Ltr)"`                                                                                                     // Example: "Ltr"
	FuelType       string  `jsonapi:"fuelType,omitempty" valid:"optional,in(e85|ron91|ron95_e5|ron95_e10|ron98|ron98_e5|ron100|diesel|diesel_gtl|diesel_b7|lpg|cng|h2|Truck Diesel|AdBlue)"` // Example: "ron95_e10"
	PricePerUnit   float32 `jsonapi:"pricePerUnit,omitempty" valid:"optional"`                                                                                                               // Example: "1.379"
	ProductName    string  `jsonapi:"productName,omitempty" valid:"optional"`                                                                                                                // Example: "Super E10"
}

// FuelPriceResponse ...
type FuelPriceResponse struct {
	ID             string  `jsonapi:"primary,fuelPrice,omitempty" valid:"optional"`                                                                                                               // Fuel Price ID
	Currency       string  `jsonapi:"attr,currency,omitempty" valid:"optional,in(EUR)"`                                                                                                           // Example: "EUR"
	FuelAmountUnit string  `jsonapi:"attr,fuelAmountUnit,omitempty" valid:"optional,in(Ltr)"`                                                                                                     // Example: "Ltr"
	FuelType       string  `jsonapi:"attr,fuelType,omitempty" valid:"optional,in(e85|ron91|ron95_e5|ron95_e10|ron98|ron98_e5|ron100|diesel|diesel_gtl|diesel_b7|lpg|cng|h2|Truck Diesel|AdBlue)"` // Example: "ron95_e10"
	PricePerUnit   float32 `jsonapi:"attr,pricePerUnit,omitempty" valid:"optional"`                                                                                                               // Example: "1.379"
	ProductName    string  `jsonapi:"attr,productName,omitempty" valid:"optional"`                                                                                                                // Example: "Super E10"
}

// GasStationResponse ...
type GasStationResponse []struct {
	ID             string                            `jsonapi:"primary,gasStation,omitempty" valid:"uuid,optional"` // Gas Station ID
	Address        *GasStationResponseAddress        `jsonapi:"attr,address,omitempty" valid:"optional"`
	Amenities      []string                          `jsonapi:"attr,amenities,omitempty" valid:"optional"` // Example: "[restaurant]"
	Latitude       float32                           `jsonapi:"attr,latitude,omitempty" valid:"optional"`  // Example: "49.013"
	Longitude      float32                           `jsonapi:"attr,longitude,omitempty" valid:"optional"` // Example: "8.425"
	OpeningHours   []*GasStationResponseOpeningHours `jsonapi:"attr,openingHours,omitempty" valid:"optional"`
	PaymentMethods []string                          `jsonapi:"attr,paymentMethods,omitempty" valid:"optional"` // Example: "[sepaDirectDebit]"
	StationName    string                            `jsonapi:"attr,stationName,omitempty" valid:"optional"`    // Example: "PACE Station"
}

// GasStationResponseAddress ...
type GasStationResponseAddress struct {
	City        string `jsonapi:"city,omitempty" valid:"optional"`        // Example: "Karlsruhe"
	CountryCode string `jsonapi:"countryCode,omitempty" valid:"optional"` // Country code in as specified in ISO 3166-1.
	HouseNo     string `jsonapi:"houseNo,omitempty" valid:"optional"`     // Example: "18"
	PostalCode  string `jsonapi:"postalCode,omitempty" valid:"optional"`  // Example: "76131"
	Street      string `jsonapi:"street,omitempty" valid:"optional"`      // Example: "Haid-und-Neu-Str."
}

// GasStationResponseOpeningHours ...
type GasStationResponseOpeningHours struct {
	OpenFromTo []string `jsonapi:"openFromTo,omitempty" valid:"optional"` // Example: "[07:30 20:30]"
	Weekdays   []string `jsonapi:"weekdays,omitempty" valid:"optional"`   // Example: "[Montag Dienstag]"
}

// PaymentMethod ...
type PaymentMethod struct {
	Attributes *PaymentMethodAttributes `jsonapi:"attributes,omitempty" valid:"optional"`
	Id         string                   `jsonapi:"id,omitempty" valid:"optional"` // Payment Method ID
	Type       string                   `jsonapi:"type,omitempty" valid:"optional,in(paymentMethod)"`
}

// PaymentMethodAttributes ...
type PaymentMethodAttributes struct {
	Kind string `jsonapi:"kind,omitempty" valid:"optional"` // Example: "sepa"
}

// PaymentMethodResponse ...
type PaymentMethodResponse struct {
	ID   string `jsonapi:"primary,paymentMethod,omitempty" valid:"optional"` // Payment Method ID
	Kind string `jsonapi:"attr,kind,omitempty" valid:"optional"`             // Example: "sepa"
}

// Pump ...
type Pump struct {
	Attributes *PumpAttributes `jsonapi:"attributes,omitempty" valid:"optional"`
	Id         string          `jsonapi:"id,omitempty" valid:"optional,uuid"`       // Pump ID
	Type       string          `jsonapi:"type,omitempty" valid:"optional,in(pump)"` // Type
}

// PumpAttributes ...
type PumpAttributes struct {
	Identifier string `jsonapi:"identifier,omitempty" valid:"optional"`                                  // Pump identifier
	Status     string `jsonapi:"status,omitempty" valid:"optional,in(free|inUse|readyToPay|outOfOrder)"` // Current pump status
}

// PumpReadyForPaymentResponse ...
type PumpReadyForPaymentResponse struct{}

// PumpResponse ...
type PumpResponse struct {
	ID         string `jsonapi:"primary,pump,omitempty" valid:"uuid,optional"`                                // Pump ID
	Identifier string `jsonapi:"attr,identifier,omitempty" valid:"optional"`                                  // Pump identifier
	Status     string `jsonapi:"attr,status,omitempty" valid:"optional,in(free|inUse|readyToPay|outOfOrder)"` // Current pump status
}

// PumpStatus Current pump status
type PumpStatus struct{}

// TransactionRequest ...
type TransactionRequest struct {
	ID              string `jsonapi:"primary,transaction,omitempty" valid:"uuid,optional"`  // Transaction ID
	MileageInMeters int64  `jsonapi:"attr,mileageInMeters,omitempty" valid:"required"`      // Example: "66435"
	PaymentMethodId string `jsonapi:"attr,paymentMethodId,omitempty" valid:"required,uuid"` // Example: "f106ac99-213c-4cf7-8c1b-1e841516026b"
	Vin             string `jsonapi:"attr,vin,omitempty" valid:"required"`                  // Example: "1B3EL46R36N102271"
}

// TransactionWithPriceCheckRequest ...
type TransactionWithPriceCheckRequest struct{}

// currency ...
type currency struct{}

// fuelAmountUnit ...
type fuelAmountUnit struct{}

/*
GetGasStationFuelingAppIdApproachingHandler handles request/response marshaling and validation for
 Get /gas-station/{fuelingAppId}/approaching
*/
func GetGasStationFuelingAppIdApproachingHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := getGasStationFuelingAppIdApproachingResponseWriter{
			ResponseWriter: w,
		}
		request := GetGasStationFuelingAppIdApproachingRequest{
			Request: r,
		}
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamFuelingAppId,
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
		err := service.GetGasStationFuelingAppIdApproaching(r.Context(), &writer, &request)
		if err != nil {
			runtime.WriteError(w, http.StatusInternalServerError, err)
		}
	})
}

/*
GetGasStationFuelingAppIdPumpsPumpIdHandler handles request/response marshaling and validation for
 Get /gas-station/{fuelingAppId}/pumps/{pumpId}
*/
func GetGasStationFuelingAppIdPumpsPumpIdHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := getGasStationFuelingAppIdPumpsPumpIdResponseWriter{
			ResponseWriter: w,
		}
		request := GetGasStationFuelingAppIdPumpsPumpIdRequest{
			Request: r,
		}
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamFuelingAppId,
			Location: runtime.ScanInPath,
			Input:    vars["fuelingAppId"],
			Name:     "fuelingAppId",
		}, &runtime.ScanParameter{
			Data:     &request.ParamPumpId,
			Location: runtime.ScanInPath,
			Input:    vars["pumpId"],
			Name:     "pumpId",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}
		err := service.GetGasStationFuelingAppIdPumpsPumpId(r.Context(), &writer, &request)
		if err != nil {
			runtime.WriteError(w, http.StatusInternalServerError, err)
		}
	})
}

/*
PostGasStationFuelingAppIdPumpsPumpIdPayHandler handles request/response marshaling and validation for
 Post /gas-station/{fuelingAppId}/pumps/{pumpId}/pay
*/
func PostGasStationFuelingAppIdPumpsPumpIdPayHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := postGasStationFuelingAppIdPumpsPumpIdPayResponseWriter{
			ResponseWriter: w,
		}
		request := PostGasStationFuelingAppIdPumpsPumpIdPayRequest{
			Request: r,
		}
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamFuelingAppId,
			Location: runtime.ScanInPath,
			Input:    vars["fuelingAppId"],
			Name:     "fuelingAppId",
		}, &runtime.ScanParameter{
			Data:     &request.ParamPumpId,
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
			err := service.PostGasStationFuelingAppIdPumpsPumpIdPay(r.Context(), &writer, &request)
			if err != nil {
				runtime.WriteError(w, http.StatusInternalServerError, err)
			}
		}
	})
}

/*
GetGasStationFuelingAppIdPumpsPumpIdWaitForStatusChangeHandler handles request/response marshaling and validation for
 Get /gas-station/{fuelingAppId}/pumps/{pumpId}/waitForStatusChange
*/
func GetGasStationFuelingAppIdPumpsPumpIdWaitForStatusChangeHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := getGasStationFuelingAppIdPumpsPumpIdWaitForStatusChangeResponseWriter{
			ResponseWriter: w,
		}
		request := GetGasStationFuelingAppIdPumpsPumpIdWaitForStatusChangeRequest{
			Request: r,
		}
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamFuelingAppId,
			Location: runtime.ScanInPath,
			Input:    vars["fuelingAppId"],
			Name:     "fuelingAppId",
		}, &runtime.ScanParameter{
			Data:     &request.ParamPumpId,
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
		err := service.GetGasStationFuelingAppIdPumpsPumpIdWaitForStatusChange(r.Context(), &writer, &request)
		if err != nil {
			runtime.WriteError(w, http.StatusInternalServerError, err)
		}
	})
}

/*
GetGasStationFuelingAppIdApproachingResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetGasStationFuelingAppIdApproachingResponseWriter interface {
	http.ResponseWriter
	OK(*GasStationResponse)
	NotFound(error)
}
type getGasStationFuelingAppIdApproachingResponseWriter struct {
	http.ResponseWriter
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *getGasStationFuelingAppIdApproachingResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *getGasStationFuelingAppIdApproachingResponseWriter) OK(data *GasStationResponse) {
	runtime.Marshal(w, data, 200)
}

/*
GetGasStationFuelingAppIdApproachingResponseWriter is a standard http.Request extended with the
un-marshaled content object
*/
type GetGasStationFuelingAppIdApproachingRequest struct {
	Request                     *http.Request `valid:"-"`
	ParamFuelingAppId           string        `valid:"required,uuid"`
	ParamExpectedAmountInLiters float32       `valid:"required"`
	ParamCarFuelType            string        `valid:"required,in(e85|ron91|ron95_e5|ron95_e10|ron98|ron98_e5|ron100|diesel|diesel_gtl|diesel_b7|lpg|cng|h2|Truck Diesel|AdBlue)"`
}

// GetGasStationFuelingAppIdPumpsPumpIdOK ...
type GetGasStationFuelingAppIdPumpsPumpIdOK struct{}

/*
GetGasStationFuelingAppIdPumpsPumpIdResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetGasStationFuelingAppIdPumpsPumpIdResponseWriter interface {
	http.ResponseWriter
	OK(*GetGasStationFuelingAppIdPumpsPumpIdOK)
	NotFound(error)
}
type getGasStationFuelingAppIdPumpsPumpIdResponseWriter struct {
	http.ResponseWriter
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *getGasStationFuelingAppIdPumpsPumpIdResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *getGasStationFuelingAppIdPumpsPumpIdResponseWriter) OK(data *GetGasStationFuelingAppIdPumpsPumpIdOK) {
	runtime.Marshal(w, data, 200)
}

/*
GetGasStationFuelingAppIdPumpsPumpIdResponseWriter is a standard http.Request extended with the
un-marshaled content object
*/
type GetGasStationFuelingAppIdPumpsPumpIdRequest struct {
	Request           *http.Request `valid:"-"`
	ParamFuelingAppId string        `valid:"required,uuid"`
	ParamPumpId       string        `valid:"required,uuid"`
}

// PostGasStationFuelingAppIdPumpsPumpIdPayCreated ...
type PostGasStationFuelingAppIdPumpsPumpIdPayCreated struct {
	ID                string                                              `jsonapi:"primary,transaction,omitempty" valid:"optional"` // Transaction ID
	VAT               *PostGasStationFuelingAppIdPumpsPumpIdPayCreatedVAT `jsonapi:"attr,VAT,omitempty" valid:"optional"`
	Currency          string                                              `jsonapi:"attr,currency,omitempty" valid:"optional,in(EUR)"`     // Example: "EUR"
	FuelingAppId      string                                              `jsonapi:"attr,fuelingAppId,omitempty" valid:"optional,uuid"`    // Example: "c30bce97-b732-4390-af38-1ac6b017aa4c"
	MileageInMeters   int64                                               `jsonapi:"attr,mileageInMeters,omitempty" valid:"optional"`      // Example: "66435"
	PaymentMethodId   string                                              `jsonapi:"attr,paymentMethodId,omitempty" valid:"optional,uuid"` // Example: "f106ac99-213c-4cf7-8c1b-1e841516026b"
	PriceIncludingVAT float32                                             `jsonapi:"attr,priceIncludingVAT,omitempty" valid:"optional"`    // Example: "69.34"
	PriceWithoutVAT   float32                                             `jsonapi:"attr,priceWithoutVAT,omitempty" valid:"optional"`      // Example: "58.27"
	PumpId            string                                              `jsonapi:"attr,pumpId,omitempty" valid:"optional,uuid"`          // Example: "460ffaad-a3c1-4199-b69e-63949ccda82f"
	Vin               string                                              `jsonapi:"attr,vin,omitempty" valid:"optional"`                  // Example: "1B3EL46R36N102271"
}

// PostGasStationFuelingAppIdPumpsPumpIdPayCreatedVAT ...
type PostGasStationFuelingAppIdPumpsPumpIdPayCreatedVAT struct {
	Amount float32 `jsonapi:"amount,omitempty" valid:"optional"` // Example: "11.07"
	Rate   float32 `jsonapi:"rate,omitempty" valid:"optional"`   // Example: "0.19"
}

/*
PostGasStationFuelingAppIdPumpsPumpIdPayResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type PostGasStationFuelingAppIdPumpsPumpIdPayResponseWriter interface {
	http.ResponseWriter
	Created(*PostGasStationFuelingAppIdPumpsPumpIdPayCreated)
	BadRequest(error)
	NotFound(error)
	Conflict(error)
}
type postGasStationFuelingAppIdPumpsPumpIdPayResponseWriter struct {
	http.ResponseWriter
}

// Conflict responds with jsonapi error (HTTP code 409)
func (w *postGasStationFuelingAppIdPumpsPumpIdPayResponseWriter) Conflict(err error) {
	runtime.WriteError(w, 409, err)
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *postGasStationFuelingAppIdPumpsPumpIdPayResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *postGasStationFuelingAppIdPumpsPumpIdPayResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// Created responds with jsonapi marshaled data (HTTP code 201)
func (w *postGasStationFuelingAppIdPumpsPumpIdPayResponseWriter) Created(data *PostGasStationFuelingAppIdPumpsPumpIdPayCreated) {
	runtime.Marshal(w, data, 201)
}

// PostGasStationFuelingAppIdPumpsPumpIdPayContent ...
type PostGasStationFuelingAppIdPumpsPumpIdPayContent struct{}

// PostGasStationFuelingAppIdPumpsPumpIdPayRequest ...
type PostGasStationFuelingAppIdPumpsPumpIdPayRequest struct {
	Request           *http.Request                                    `valid:"-"`
	Content           *PostGasStationFuelingAppIdPumpsPumpIdPayContent `valid:"-"`
	ParamFuelingAppId string                                           `valid:"required,uuid"`
	ParamPumpId       string                                           `valid:"required,uuid"`
}

// GetGasStationFuelingAppIdPumpsPumpIdWaitForStatusChangeOK ...
type GetGasStationFuelingAppIdPumpsPumpIdWaitForStatusChangeOK struct{}

/*
GetGasStationFuelingAppIdPumpsPumpIdWaitForStatusChangeResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetGasStationFuelingAppIdPumpsPumpIdWaitForStatusChangeResponseWriter interface {
	http.ResponseWriter
	OK(*GetGasStationFuelingAppIdPumpsPumpIdWaitForStatusChangeOK)
	NotFound(error)
	RequestTimeout(error)
}
type getGasStationFuelingAppIdPumpsPumpIdWaitForStatusChangeResponseWriter struct {
	http.ResponseWriter
}

// RequestTimeout responds with jsonapi error (HTTP code 408)
func (w *getGasStationFuelingAppIdPumpsPumpIdWaitForStatusChangeResponseWriter) RequestTimeout(err error) {
	runtime.WriteError(w, 408, err)
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *getGasStationFuelingAppIdPumpsPumpIdWaitForStatusChangeResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *getGasStationFuelingAppIdPumpsPumpIdWaitForStatusChangeResponseWriter) OK(data *GetGasStationFuelingAppIdPumpsPumpIdWaitForStatusChangeOK) {
	runtime.Marshal(w, data, 200)
}

/*
GetGasStationFuelingAppIdPumpsPumpIdWaitForStatusChangeResponseWriter is a standard http.Request extended with the
un-marshaled content object
*/
type GetGasStationFuelingAppIdPumpsPumpIdWaitForStatusChangeRequest struct {
	Request           *http.Request `valid:"-"`
	ParamFuelingAppId string        `valid:"required,uuid"`
	ParamPumpId       string        `valid:"required,uuid"`
	ParamLastStatus   string        `valid:"optional,in(free|inUse|readyToPay|outOfOrder)"`
	ParamTimeout      int64         `valid:"optional"`
}
type Service interface {
	/*
	   GetGasStationFuelingAppIdApproaching Gather information when approaching at the forecourt


	   This request will:
	   * Return a list of available paymentMethodIds
	   * Return up-to-date price information (price structure) at the gas station
	   * Return a list of pumps available at the gas station together with the current status (free, inUse, readyToPay, outOfOrder)
	   * Create payment tokens for all paymentMethods of the user and pre-authorise the calculated maximum amount of money (background task)
	*/
	GetGasStationFuelingAppIdApproaching(context.Context, GetGasStationFuelingAppIdApproachingResponseWriter, *GetGasStationFuelingAppIdApproachingRequest) error
	/*
	   GetGasStationFuelingAppIdPumpsPumpId Return current pump information

	   Returns the current pump status (free, inUse, readyToPay, outOfOrder) and identifier. If the status is readyToPay, the result also contains fuelType, productName, fuelAmount, fuelAmountUnit, VAT (amount & rate), priceWithoutVAT, priceIncludingVAT, currency.
	*/
	GetGasStationFuelingAppIdPumpsPumpId(context.Context, GetGasStationFuelingAppIdPumpsPumpIdResponseWriter, *GetGasStationFuelingAppIdPumpsPumpIdRequest) error
	/*
	   PostGasStationFuelingAppIdPumpsPumpIdPay Process payment

	   Process payment and notify user if transaction is finished successfully. You can optionally provide `priceIncludingVAT`and `currency` in the request body to check if the price the user has seen is still correct.
	*/
	PostGasStationFuelingAppIdPumpsPumpIdPay(context.Context, PostGasStationFuelingAppIdPumpsPumpIdPayResponseWriter, *PostGasStationFuelingAppIdPumpsPumpIdPayRequest) error
	/*
	   GetGasStationFuelingAppIdPumpsPumpIdWaitForStatusChange Wait for a status change on a given pump

	   Uses **long polling** to wait for a status change on a given pump. Returns as soon as the status has changed or after the number of seconds provided by the optional `timeout` query parameter (default timeout is 30 seconds). In case of timeout (408 status code) you're safe to start the request again. Instantaneously returns if `lastStatus` was given and already changed between request. If successful, it returns the same structure as the normal status call
	*/
	GetGasStationFuelingAppIdPumpsPumpIdWaitForStatusChange(context.Context, GetGasStationFuelingAppIdPumpsPumpIdWaitForStatusChangeResponseWriter, *GetGasStationFuelingAppIdPumpsPumpIdWaitForStatusChangeRequest) error
}

/*
Router implements: PACE Fueling API

Fueling API
*/
func Router(service Service) *mux.Router {
	router := mux.NewRouter()
	// Subrouter s1 - https://api.pace.cloud/fueling/beta
	s1 := router.PathPrefix("/fueling/beta").Subrouter()
	s1.Methods("POST").Path("/gas-station/{fuelingAppId}/pumps/{pumpId}/pay").Handler(PostGasStationFuelingAppIdPumpsPumpIdPayHandler(service)).Name("PostGasStationFuelingAppIdPumpsPumpIdPay")
	s1.Methods("GET").Path("/gas-station/{fuelingAppId}/pumps/{pumpId}/waitForStatusChange").Handler(GetGasStationFuelingAppIdPumpsPumpIdWaitForStatusChangeHandler(service)).Name("GetGasStationFuelingAppIdPumpsPumpIdWaitForStatusChange")
	s1.Methods("GET").Path("/gas-station/{fuelingAppId}/pumps/{pumpId}").Handler(GetGasStationFuelingAppIdPumpsPumpIdHandler(service)).Name("GetGasStationFuelingAppIdPumpsPumpId")
	s1.Methods("GET").Path("/gas-station/{fuelingAppId}/approaching").Handler(GetGasStationFuelingAppIdApproachingHandler(service)).Name("GetGasStationFuelingAppIdApproaching")
	return router
}
