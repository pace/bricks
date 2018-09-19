package poi

import (
	"context"
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
	ID             string          `jsonapi:"primary,fuelPrice,omitempty" valid:"uuid,optional"` // Fuel Price ID
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
	ID                string                            `jsonapi:"primary,gasStation,omitempty" valid:"uuid,optional"` // Gas Station ID
	Address           *GasStationResponseAddress        `json:"address,omitempty" jsonapi:"attr,address,omitempty" valid:"optional"`
	Amenities         []string                          `json:"amenities,omitempty" jsonapi:"attr,amenities,omitempty" valid:"optional"` // Example: "[restaurant]"
	Latitude          float32                           `json:"latitude,omitempty" jsonapi:"attr,latitude,omitempty" valid:"optional"`   // Example: "49.013"
	Longitude         float32                           `json:"longitude,omitempty" jsonapi:"attr,longitude,omitempty" valid:"optional"` // Example: "8.425"
	OpeningHours      []*GasStationResponseOpeningHours `json:"openingHours,omitempty" jsonapi:"attr,openingHours,omitempty" valid:"optional"`
	PaymentMethods    []string                          `json:"paymentMethods,omitempty" jsonapi:"attr,paymentMethods,omitempty" valid:"optional"` // Example: "[sepa]"
	StationName       string                            `json:"stationName,omitempty" jsonapi:"attr,stationName,omitempty" valid:"optional"`       // Example: "PACE Station"
	FuelPrices        []*FuelPrice                      `json:"fuelPrices,omitempty" jsonapi:"attr,fuelPrices,omitempty" valid:"optional"`
	LocationBasedApps []*LocationBasedApp               `json:"locationBasedApps,omitempty" jsonapi:"attr,locationBasedApps,omitempty" valid:"optional"`
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

// LocationBasedApp ...
type LocationBasedApp struct {
	ID                   string      `jsonapi:"primary,locationBasedApp,omitempty" valid:"uuid,optional"`                                   // Location-based app ID
	AndroidInstantAppURL string      `json:"androidInstantAppUrl,omitempty" jsonapi:"attr,androidInstantAppUrl,omitempty" valid:"optional"` // Android instant app URL
	AppArea              [][]float32 `json:"appArea,omitempty" jsonapi:"attr,appArea,omitempty" valid:"optional"`                           // Example: "[[49.012 8.424] [49.1 9.34] [48.7 8.92]]"
	AppType              string      `json:"appType,omitempty" jsonapi:"attr,appType,omitempty" valid:"optional"`
	InsideAppArea        bool        `json:"insideAppArea,omitempty" jsonapi:"attr,insideAppArea,omitempty" valid:"optional"` // Boolean flag if the current position is inside the app area (polygon).
	LogoURL              string      `json:"logoUrl,omitempty" jsonapi:"attr,logoUrl,omitempty" valid:"optional"`             // Logo URL
	PwaURL               string      `json:"pwaUrl,omitempty" jsonapi:"attr,pwaUrl,omitempty" valid:"optional"`               // Progressive web application URL
	Subtitle             string      `json:"subtitle,omitempty" jsonapi:"attr,subtitle,omitempty" valid:"optional"`           // Example: "Zahle bargeldlos mit der PACE Fueling App"
	Title                string      `json:"title,omitempty" jsonapi:"attr,title,omitempty" valid:"optional"`                 // Example: "PACE Fueling App"
}

// LocationBasedAppResponse ...
type LocationBasedAppResponse *LocationBasedApp

// LocationBasedAppsResponseItem ...
type LocationBasedAppsResponseItem struct {
	ID                   string      `jsonapi:"primary,locationBasedApp,omitempty" valid:"uuid,optional"`                                   // Location-based app ID
	AndroidInstantAppURL string      `json:"androidInstantAppUrl,omitempty" jsonapi:"attr,androidInstantAppUrl,omitempty" valid:"optional"` // Android instant app URL
	AppArea              [][]float32 `json:"appArea,omitempty" jsonapi:"attr,appArea,omitempty" valid:"optional"`                           // Example: "[[49.012 8.424] [49.1 9.34] [49.012 8.424]]"
	AppType              string      `json:"appType,omitempty" jsonapi:"attr,appType,omitempty" valid:"optional"`
	InsideAppArea        bool        `json:"insideAppArea,omitempty" jsonapi:"attr,insideAppArea,omitempty" valid:"optional"` // Boolean flag if the current position is inside the app area (polygon).
	LogoURL              string      `json:"logoUrl,omitempty" jsonapi:"attr,logoUrl,omitempty" valid:"optional"`             // Logo URL
	PwaURL               string      `json:"pwaUrl,omitempty" jsonapi:"attr,pwaUrl,omitempty" valid:"optional"`               // Progressive web application URL
	Subtitle             string      `json:"subtitle,omitempty" jsonapi:"attr,subtitle,omitempty" valid:"optional"`           // Example: "Zahle bargeldlos mit der PACE Fueling App"
	Title                string      `json:"title,omitempty" jsonapi:"attr,title,omitempty" valid:"optional"`                 // Example: "PACE Fueling App"
}

// LocationBasedAppsResponse ...
type LocationBasedAppsResponse []*LocationBasedAppsResponseItem

// Currency ...
type Currency string

// FuelAmountUnit ...
type FuelAmountUnit string

/*
CheckForPaceAppHandler handles request/response marshaling and validation for
 Get /beta/check-for-pace-app
*/
func CheckForPaceAppHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer func() {
			if rp := recover(); rp != nil {
				log.Ctx(ctx).Error().Str("handler", "CheckForPaceAppHandler").Msgf("Panic: %v", rp)
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
		handlerSpan = opentracing.StartSpan("CheckForPaceAppHandler", opentracing.ChildOf(wireContext))
		handlerSpan.LogFields(olog.String("req_id", log.RequestID(r)))
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		ctx = opentracing.ContextWithSpan(r.Context(), handlerSpan)
		writer := checkForPaceAppResponseWriter{
			ResponseWriter: jsonapimetrics.NewMetric("poi", "/beta/check-for-pace-app", w, r),
		}
		request := CheckForPaceAppRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamLatitude,
			Location: runtime.ScanInQuery,
			Input:    vars["latitude"],
			Name:     "latitude",
		}, &runtime.ScanParameter{
			Data:     &request.ParamLongitude,
			Location: runtime.ScanInQuery,
			Input:    vars["longitude"],
			Name:     "longitude",
		}, &runtime.ScanParameter{
			Data:     &request.ParamGpsSource,
			Location: runtime.ScanInQuery,
			Input:    vars["gpsSource"],
			Name:     "gpsSource",
		}, &runtime.ScanParameter{
			Data:     &request.ParamAppType,
			Location: runtime.ScanInQuery,
			Input:    vars["appType"],
			Name:     "appType",
		}, &runtime.ScanParameter{
			Data:     &request.ParamAccuracy,
			Location: runtime.ScanInQuery,
			Input:    vars["accuracy"],
			Name:     "accuracy",
		}, &runtime.ScanParameter{
			Data:     &request.ParamDeviation,
			Location: runtime.ScanInQuery,
			Input:    vars["deviation"],
			Name:     "deviation",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Invoke service that implements the business logic
		err = service.CheckForPaceApp(ctx, &writer, &request)
		if err != nil {
			runtime.WriteError(w, http.StatusInternalServerError, err)
		}
	})
}

/*
SearchHandler handles request/response marshaling and validation for
 Get /beta/search
*/
func SearchHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer func() {
			if rp := recover(); rp != nil {
				log.Ctx(ctx).Error().Str("handler", "SearchHandler").Msgf("Panic: %v", rp)
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
		handlerSpan = opentracing.StartSpan("SearchHandler", opentracing.ChildOf(wireContext))
		handlerSpan.LogFields(olog.String("req_id", log.RequestID(r)))
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		ctx = opentracing.ContextWithSpan(r.Context(), handlerSpan)
		writer := searchResponseWriter{
			ResponseWriter: jsonapimetrics.NewMetric("poi", "/beta/search", w, r),
		}
		request := SearchRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamPoiType,
			Location: runtime.ScanInQuery,
			Input:    vars["poiType"],
			Name:     "poiType",
		}, &runtime.ScanParameter{
			Data:     &request.ParamAppType,
			Location: runtime.ScanInQuery,
			Input:    vars["appType"],
			Name:     "appType",
		}, &runtime.ScanParameter{
			Data:     &request.ParamGpsSource,
			Location: runtime.ScanInQuery,
			Input:    vars["gpsSource"],
			Name:     "gpsSource",
		}, &runtime.ScanParameter{
			Data:     &request.ParamInclude,
			Location: runtime.ScanInQuery,
			Input:    vars["include"],
			Name:     "include",
		}, &runtime.ScanParameter{
			Data:     &request.ParamLatitude,
			Location: runtime.ScanInQuery,
			Input:    vars["latitude"],
			Name:     "latitude",
		}, &runtime.ScanParameter{
			Data:     &request.ParamLongitude,
			Location: runtime.ScanInQuery,
			Input:    vars["longitude"],
			Name:     "longitude",
		}, &runtime.ScanParameter{
			Data:     &request.ParamRadius,
			Location: runtime.ScanInQuery,
			Input:    vars["radius"],
			Name:     "radius",
		}, &runtime.ScanParameter{
			Data:     &request.ParamAccuracy,
			Location: runtime.ScanInQuery,
			Input:    vars["accuracy"],
			Name:     "accuracy",
		}, &runtime.ScanParameter{
			Data:     &request.ParamDeviation,
			Location: runtime.ScanInQuery,
			Input:    vars["deviation"],
			Name:     "deviation",
		}, &runtime.ScanParameter{
			Data:     &request.ParamBoundingBox,
			Location: runtime.ScanInQuery,
			Input:    vars["boundingBox"],
			Name:     "boundingBox",
		}, &runtime.ScanParameter{
			Data:     &request.ParamPath,
			Location: runtime.ScanInQuery,
			Input:    vars["path"],
			Name:     "path",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Invoke service that implements the business logic
		err = service.Search(ctx, &writer, &request)
		if err != nil {
			runtime.WriteError(w, http.StatusInternalServerError, err)
		}
	})
}

/*
CheckForPaceAppResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type CheckForPaceAppResponseWriter interface {
	http.ResponseWriter
	OK(LocationBasedAppsResponse)
	BadRequest(error)
}
type checkForPaceAppResponseWriter struct {
	http.ResponseWriter
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *checkForPaceAppResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *checkForPaceAppResponseWriter) OK(data LocationBasedAppsResponse) {
	runtime.Marshal(w, data, 200)
}

/*
CheckForPaceAppRequest is a standard http.Request extended with the
un-marshaled content object
*/
type CheckForPaceAppRequest struct {
	Request        *http.Request `valid:"-"`
	ParamLatitude  float32       `valid:"required"`
	ParamLongitude float32       `valid:"required"`
	ParamGpsSource string        `valid:"required,in(raw|mapMatched)"`
	ParamAppType   string        `valid:"required,in(fueling)"`
	ParamAccuracy  float32       `valid:"optional"`
	ParamDeviation float32       `valid:"optional"`
}

/*
SearchResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type SearchResponseWriter interface {
	http.ResponseWriter
	OK(GasStationResponse)
	BadRequest(error)
}
type searchResponseWriter struct {
	http.ResponseWriter
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *searchResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *searchResponseWriter) OK(data GasStationResponse) {
	runtime.Marshal(w, data, 200)
}

/*
SearchRequest is a standard http.Request extended with the
un-marshaled content object
*/
type SearchRequest struct {
	Request          *http.Request `valid:"-"`
	ParamPoiType     string        `valid:"required,in(gasStation)"`
	ParamAppType     []string      `valid:"required,in(fueling)"`
	ParamGpsSource   string        `valid:"required,in(raw|mapMatched)"`
	ParamInclude     string        `valid:"required,in(insideAppArea)"`
	ParamLatitude    float32       `valid:"optional"`
	ParamLongitude   float32       `valid:"optional"`
	ParamRadius      float32       `valid:"optional"`
	ParamAccuracy    float32       `valid:"optional"`
	ParamDeviation   float32       `valid:"optional"`
	ParamBoundingBox []float32     `valid:"optional"`
	ParamPath        [][]float32   `valid:"optional"`
}

// Service interface for all handlers
type Service interface {
	/*
	   CheckForPaceApp Get location-based apps


	   These location-based PACE apps deliver additional services for PACE customers based on their current position.
	   You can (or should) trigger this whenever:
	   * A longer stand-still is detected
	   * The engine is turned off
	   * Every 5 seconds if the user "left the road"

	   Please note that calling this API is very cheap and can be done regularly.
	*/
	CheckForPaceApp(context.Context, CheckForPaceAppResponseWriter, *CheckForPaceAppRequest) error
	/*
	   Search Search for gas stations

	   There are three possibilities to search for gas stations. If you want to search in a specific radius around a given longitude and latitude you have to provide the following query parameters:

	   * latitude (required)
	   * longitude (required)
	   * radius (required)
	   * accuracy (optional)

	   If you want to search in a given bounding box you have to provide the following query parameters:

	   * boundingBox (required)

	   If you want to search along a given path you have to provide the following query parameters:
	   * path (required)
	   * radius (required)

	   If you have map-matched GPS data you can also provide a `deviation` query parameter. By using this query type, the evaluation if the user is inside the polygon of a specific location-based PACE app needs to be done by the client.
	*/
	Search(context.Context, SearchResponseWriter, *SearchRequest) error
}

/*
Router implements: PACE POI API

POI API
*/
func Router(service Service) *mux.Router {
	router := mux.NewRouter()
	// Subrouter s1 - https://api.pace.cloud/poi
	s1 := router.PathPrefix("/poi").Subrouter()
	s1.Methods("GET").Path("/beta/check-for-pace-app").Handler(CheckForPaceAppHandler(service)).Name("CheckForPaceApp")
	s1.Methods("GET").Path("/beta/search").Handler(SearchHandler(service)).Name("Search")
	return router
}
