package poi

import (
	"context"
	mux "github.com/gorilla/mux"
	runtime "lab.jamit.de/pace/web/libs/go-microservice/http/jsonapi/runtime"
	"net/http"
)

// FuelPrice ...
type FuelPrice struct {
	Attributes *FuelPriceAttributes `jsonapi:"attributes,omitempty" valid:"optional"`
	ID         string               `jsonapi:"id,omitempty" valid:"optional"`                 // Fuel Price ID
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

// LocationBasedApp ...
type LocationBasedApp struct {
	Attributes *LocationBasedAppAttributes `jsonapi:"attributes,omitempty" valid:"optional"`
	ID         string                      `jsonapi:"id,omitempty" valid:"optional,uuid"`                   // Location-based app ID
	Type       string                      `jsonapi:"type,omitempty" valid:"optional,in(locationBasedApp)"` // Type
}

// LocationBasedAppAttributes ...
type LocationBasedAppAttributes struct {
	AndroidInstantAppURL string      `jsonapi:"androidInstantAppUrl,omitempty" valid:"optional"` // Android instant app URL
	AppArea              [][]float32 `jsonapi:"appArea,omitempty" valid:"optional"`              // Example: "[[49.012 8.424] [49.1 9.34] [48.7 8.92]]"
	AppType              string      `jsonapi:"appType,omitempty" valid:"optional,in(fueling)"`
	InsideAppArea        bool        `jsonapi:"insideAppArea,omitempty" valid:"optional"` // Boolean flag if the current position is inside the app area (polygon).
	LogoURL              string      `jsonapi:"logoUrl,omitempty" valid:"optional"`       // Logo URL
	PwaURL               string      `jsonapi:"pwaUrl,omitempty" valid:"optional"`        // Progressive web application URL
	Subtitle             string      `jsonapi:"subtitle,omitempty" valid:"optional"`      // Example: "Zahle bargeldlos mit der PACE Fueling App"
	Title                string      `jsonapi:"title,omitempty" valid:"optional"`         // Example: "PACE Fueling App"
}

// LocationBasedAppResponse ...
type LocationBasedAppResponse struct {
	ID                   string      `jsonapi:"primary,locationBasedApp,omitempty" valid:"uuid,optional"` // Location-based app ID
	AndroidInstantAppURL string      `jsonapi:"attr,androidInstantAppUrl,omitempty" valid:"optional"`     // Android instant app URL
	AppArea              [][]float32 `jsonapi:"attr,appArea,omitempty" valid:"optional"`                  // Example: "[[49.012 8.424] [49.1 9.34] [48.7 8.92]]"
	AppType              string      `jsonapi:"attr,appType,omitempty" valid:"optional,in(fueling)"`
	InsideAppArea        bool        `jsonapi:"attr,insideAppArea,omitempty" valid:"optional"` // Boolean flag if the current position is inside the app area (polygon).
	LogoURL              string      `jsonapi:"attr,logoUrl,omitempty" valid:"optional"`       // Logo URL
	PwaURL               string      `jsonapi:"attr,pwaUrl,omitempty" valid:"optional"`        // Progressive web application URL
	Subtitle             string      `jsonapi:"attr,subtitle,omitempty" valid:"optional"`      // Example: "Zahle bargeldlos mit der PACE Fueling App"
	Title                string      `jsonapi:"attr,title,omitempty" valid:"optional"`         // Example: "PACE Fueling App"
}

// LocationBasedAppsResponse ...
type LocationBasedAppsResponse []struct {
	ID                   string      `jsonapi:"primary,locationBasedApp,omitempty" valid:"uuid,optional"` // Location-based app ID
	AndroidInstantAppURL string      `jsonapi:"attr,androidInstantAppUrl,omitempty" valid:"optional"`     // Android instant app URL
	AppArea              [][]float32 `jsonapi:"attr,appArea,omitempty" valid:"optional"`                  // Example: "[[49.012 8.424] [49.1 9.34] [49.012 8.424]]"
	AppType              string      `jsonapi:"attr,appType,omitempty" valid:"optional,in(fueling)"`
	InsideAppArea        bool        `jsonapi:"attr,insideAppArea,omitempty" valid:"optional"` // Boolean flag if the current position is inside the app area (polygon).
	LogoURL              string      `jsonapi:"attr,logoUrl,omitempty" valid:"optional"`       // Logo URL
	PwaURL               string      `jsonapi:"attr,pwaUrl,omitempty" valid:"optional"`        // Progressive web application URL
	Subtitle             string      `jsonapi:"attr,subtitle,omitempty" valid:"optional"`      // Example: "Zahle bargeldlos mit der PACE Fueling App"
	Title                string      `jsonapi:"attr,title,omitempty" valid:"optional"`         // Example: "PACE Fueling App"
}

// currency ...
type currency struct{}

// fuelAmountUnit ...
type fuelAmountUnit struct{}

/*
GetCheckForPaceAppHandler handles request/response marshaling and validation for
 Get /check-for-pace-app
*/
func GetCheckForPaceAppHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := getCheckForPaceAppResponseWriter{
			ResponseWriter: w,
		}
		request := GetCheckForPaceAppRequest{
			Request: r,
		}
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
		err := service.GetCheckForPaceApp(r.Context(), &writer, &request)
		if err != nil {
			runtime.WriteError(w, http.StatusInternalServerError, err)
		}
	})
}

/*
GetSearchHandler handles request/response marshaling and validation for
 Get /search
*/
func GetSearchHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := getSearchResponseWriter{
			ResponseWriter: w,
		}
		request := GetSearchRequest{
			Request: r,
		}
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
		err := service.GetSearch(r.Context(), &writer, &request)
		if err != nil {
			runtime.WriteError(w, http.StatusInternalServerError, err)
		}
	})
}

/*
GetCheckForPaceAppResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetCheckForPaceAppResponseWriter interface {
	http.ResponseWriter
	OK(*LocationBasedAppsResponse)
	BadRequest(error)
}
type getCheckForPaceAppResponseWriter struct {
	http.ResponseWriter
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *getCheckForPaceAppResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *getCheckForPaceAppResponseWriter) OK(data *LocationBasedAppsResponse) {
	runtime.Marshal(w, data, 200)
}

/*
GetCheckForPaceAppResponseWriter is a standard http.Request extended with the
un-marshaled content object
*/
type GetCheckForPaceAppRequest struct {
	Request        *http.Request `valid:"-"`
	ParamLatitude  float32       `valid:"required"`
	ParamLongitude float32       `valid:"required"`
	ParamGpsSource string        `valid:"required,in(raw|mapMatched)"`
	ParamAppType   string        `valid:"required,in(fueling)"`
	ParamAccuracy  float32       `valid:"optional"`
	ParamDeviation float32       `valid:"optional"`
}

/*
GetSearchResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetSearchResponseWriter interface {
	http.ResponseWriter
	OK(*GasStationResponse)
}
type getSearchResponseWriter struct {
	http.ResponseWriter
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *getSearchResponseWriter) OK(data *GasStationResponse) {
	runtime.Marshal(w, data, 200)
}

/*
GetSearchResponseWriter is a standard http.Request extended with the
un-marshaled content object
*/
type GetSearchRequest struct {
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
type Service interface {
	/*
	   GetCheckForPaceApp Get location-based apps


	   These location-based PACE apps deliver additional services for PACE customers based on their current position.
	   You can (or should) trigger this whenever:
	   * A longer stand-still is detected
	   * The engine is turned off
	   * Every 5 seconds if the user "left the road"

	   Please note that calling this API is very cheap and can be done regularly.
	*/
	GetCheckForPaceApp(context.Context, GetCheckForPaceAppResponseWriter, *GetCheckForPaceAppRequest) error
	/*
	   GetSearch Search for gas stations

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
	GetSearch(context.Context, GetSearchResponseWriter, *GetSearchRequest) error
}

/*
Router implements: PACE POI API

POI API
*/
func Router(service Service) *mux.Router {
	router := mux.NewRouter()
	// Subrouter s1 - https://api.pace.cloud/poi/beta
	s1 := router.PathPrefix("/poi/beta").Subrouter()
	s1.Methods("GET").Path("/check-for-pace-app").Handler(GetCheckForPaceAppHandler(service)).Name("GetCheckForPaceApp")
	s1.Methods("GET").Path("/search").Handler(GetSearchHandler(service)).Name("GetSearch")
	return router
}
