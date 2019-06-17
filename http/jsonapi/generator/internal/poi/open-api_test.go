// nolint
package poi

import (
	"context"
	jsonapi "github.com/google/jsonapi"
	mux "github.com/gorilla/mux"
	opentracing "github.com/opentracing/opentracing-go"
	runtime "github.com/pace/bricks/http/jsonapi/runtime"
	errors "github.com/pace/bricks/maintenance/errors"
	metrics "github.com/pace/bricks/maintenance/metric/jsonapi"
	"net/http"
	"time"
)

// AppPOIsRelationshipsItem ...
type AppPOIsRelationshipsItem struct {
	ID string `jsonapi:"primary,pois,omitempty" valid:"uuid,optional"` // ID of the for the referenced object
}

// AppPOIsRelationships ...
type AppPOIsRelationships []*AppPOIsRelationshipsItem

// CommonCountryID Country this policy applies to (as ISO3166Alpha2)
type CommonCountryID string

// CommonGeoJSONPoint https://tools.ietf.org/html/rfc7946#section-3.1.2
type CommonGeoJSONPoint struct {
	Coordinates CommonLatLong `json:"coordinates,omitempty" jsonapi:"attr,coordinates,omitempty" valid:"optional"`
	Type        string        `json:"type,omitempty" jsonapi:"attr,type,omitempty" valid:"optional,in(Point)"` // Example: "Point"
}

// CommonGeoJSONPolygon https://tools.ietf.org/html/rfc7946#section-3.1.6; used as [bounding box](https://tools.ietf.org/html/rfc7946#section-5)
type CommonGeoJSONPolygon struct {
	Coordinates []CommonLatLong `json:"coordinates,omitempty" jsonapi:"attr,coordinates,omitempty" valid:"optional"` // Example: "[[49.012 8.424] [49.1 9.34] [49.012 8.424]]"
	Type        string          `json:"type,omitempty" jsonapi:"attr,type,omitempty" valid:"optional,in(Polygon)"`   // Example: "Polygon"
}

// CommonLatLong https://tools.ietf.org/html/rfc7946
type CommonLatLong []float32

// CommonOpeningHoursTimespans ...
type CommonOpeningHoursTimespans struct {
	From string `json:"from,omitempty" jsonapi:"attr,from,omitempty" valid:"optional"` // Example: "07:30"
	To   string `json:"to,omitempty" jsonapi:"attr,to,omitempty" valid:"optional"`     // Example: "20:30"
}

// CommonOpeningHours ...
type CommonOpeningHours []struct {
	Days      []string                      `json:"days,omitempty" jsonapi:"attr,days,omitempty" valid:"optional"` // Example: "[Montag Dienstag]"
	Timespans []CommonOpeningHoursTimespans `json:"timespans,omitempty" jsonapi:"attr,timespans,omitempty" valid:"optional"`
}

// Event ...
type Event struct {
	ID        string      `jsonapi:"primary,events,omitempty" valid:"uuid,optional"` // Event ID
	CreatedAt time.Time   `json:"createdAt,omitempty" jsonapi:"attr,createdAt,omitempty,iso8601" valid:"optional"`
	EventAt   time.Time   `json:"eventAt,omitempty" jsonapi:"attr,eventAt,omitempty,iso8601" valid:"optional"`
	Fields    []FieldData `json:"fields,omitempty" jsonapi:"attr,fields,omitempty" valid:"optional"`
	UserID    string      `json:"userId,omitempty" jsonapi:"attr,userId,omitempty" valid:"optional,uuid"` // Tracks who did last change
}

// Events ...
type Events []*Event

// FieldData ...
type FieldData struct {
	Field FieldName `json:"field,omitempty" jsonapi:"attr,field,omitempty" valid:"optional"`
	Value string    `json:"value,omitempty" jsonapi:"attr,value,omitempty" valid:"optional"` // escaped json
}

// FieldMetaData ...
type FieldMetaData struct {
	SourceID  string    `json:"SourceId,omitempty" jsonapi:"attr,SourceId,omitempty" valid:"optional,uuid"` // Source ID
	UpdatedAt time.Time `json:"UpdatedAt,omitempty" jsonapi:"attr,UpdatedAt,omitempty,iso8601" valid:"optional"`
	Field     FieldName `json:"field,omitempty" jsonapi:"attr,field,omitempty" valid:"optional"`
}

// FieldName ...
type FieldName string

// FuelPrice ...
type FuelPrice struct {
	ID          string   `jsonapi:"primary,fuelPrice,omitempty" valid:"uuid,optional"` // Fuel Price ID
	Currency    Currency `json:"currency,omitempty" jsonapi:"attr,currency,omitempty" valid:"optional"`
	FuelType    string   `json:"fuelType,omitempty" jsonapi:"attr,fuelType,omitempty" valid:"optional,in(e85|ron91|ron95_e5|ron95_e10|ron98|ron98_e5|ron100|diesel|diesel_gtl|diesel_b7|lpg|lng|cng|h2|Truck Diesel|AdBlue)"` // Example: "ron95_e10"
	Price       float32  `json:"price,omitempty" jsonapi:"attr,price,omitempty" valid:"optional"`                                                                                                                             // per liter
	ProductName string   `json:"productName,omitempty" jsonapi:"attr,productName,omitempty" valid:"optional"`                                                                                                                 // Example: "Super E10"
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

// GasStation ...
type GasStation struct {
	ID                string              `jsonapi:"primary,gasStation,omitempty" valid:"uuid,optional"` // Gas Station ID
	Address           GasStationAddress   `json:"address,omitempty" jsonapi:"attr,address,omitempty" valid:"optional"`
	Amenities         []string            `json:"amenities,omitempty" jsonapi:"attr,amenities,omitempty" valid:"optional"` // Example: "[restaurant]"
	Latitude          float32             `json:"latitude,omitempty" jsonapi:"attr,latitude,omitempty" valid:"optional"`   // Example: "49.013"
	Longitude         float32             `json:"longitude,omitempty" jsonapi:"attr,longitude,omitempty" valid:"optional"` // Example: "8.425"
	OpeningHours      CommonOpeningHours  `json:"openingHours,omitempty" jsonapi:"attr,openingHours,omitempty" valid:"optional"`
	PaymentMethods    []string            `json:"paymentMethods,omitempty" jsonapi:"attr,paymentMethods,omitempty" valid:"optional,in(sepa)"` // Example: "[sepa]"
	StationName       string              `json:"stationName,omitempty" jsonapi:"attr,stationName,omitempty" valid:"optional"`                // Example: "PACE Station"
	FuelPrices        []*FuelPrice        `json:"fuelPrices,omitempty" jsonapi:"relation,fuelPrices,omitempty" valid:"optional"`
	LocationBasedApps []*LocationBasedApp `json:"locationBasedApps,omitempty" jsonapi:"relation,locationBasedApps,omitempty" valid:"optional"`
}

// GasStations ...
type GasStations []*GasStation

// LocationBasedAppMeta ...
type LocationBasedAppMeta struct {
	AppArea       CommonGeoJSONPolygon `json:"appArea,omitempty" jsonapi:"attr,appArea,omitempty" valid:"optional"`
	InsideAppArea bool                 `json:"insideAppArea,omitempty" jsonapi:"attr,insideAppArea,omitempty" valid:"optional"` // Boolean flag if the current position is inside the app area (polygon).
}

// LocationBasedApp ...
type LocationBasedApp struct {
	ID                   string                `jsonapi:"primary,locationBasedApp,omitempty" valid:"uuid,optional"`                                   // Location-based app ID
	AndroidInstantAppURL string                `json:"androidInstantAppUrl,omitempty" jsonapi:"attr,androidInstantAppUrl,omitempty" valid:"optional"` // Android instant app URL
	AppType              string                `json:"appType,omitempty" jsonapi:"attr,appType,omitempty" valid:"optional,in(fueling)"`
	LogoURL              string                `json:"logoUrl,omitempty" jsonapi:"attr,logoUrl,omitempty" valid:"optional"`   // Logo URL
	PwaURL               string                `json:"pwaUrl,omitempty" jsonapi:"attr,pwaUrl,omitempty" valid:"optional"`     // Progressive web application URL
	Subtitle             string                `json:"subtitle,omitempty" jsonapi:"attr,subtitle,omitempty" valid:"optional"` // Example: "Zahle bargeldlos mit der PACE Fueling App"
	Title                string                `json:"title,omitempty" jsonapi:"attr,title,omitempty" valid:"optional"`       // Example: "PACE Fueling App"
	Meta                 *LocationBasedAppMeta // Resource meta data (json:api meta)
}

// JSONAPIMeta implements the meta data API for json:api
func (r *LocationBasedApp) JSONAPIMeta() *jsonapi.Meta {
	if r.Meta == nil {
		return nil
	}
	meta := make(jsonapi.Meta)
	meta["appArea"] = r.Meta.AppArea
	meta["insideAppArea"] = r.Meta.InsideAppArea
	return &meta
}

// LocationBasedApps ...
type LocationBasedApps []*LocationBasedApp

// POI ...
type POI struct {
	ID         string               `jsonapi:"primary,SpeedCamera,omitempty" valid:"uuid,optional"` // POI ID
	Active     bool                 `json:"active,omitempty" jsonapi:"attr,active,omitempty" valid:"optional"`
	Boundary   CommonGeoJSONPolygon `json:"boundary,omitempty" jsonapi:"attr,boundary,omitempty" valid:"optional"`
	CountryID  CommonCountryID      `json:"countryId,omitempty" jsonapi:"attr,countryId,omitempty" valid:"optional"`
	CreatedAt  time.Time            `json:"createdAt,omitempty" jsonapi:"attr,createdAt,omitempty,iso8601" valid:"optional"`
	Data       []FieldData          `json:"data,omitempty" jsonapi:"attr,data,omitempty" valid:"optional"` // a JSON field containing POI specific data
	LastSeenAt time.Time            `json:"lastSeenAt,omitempty" jsonapi:"attr,lastSeenAt,omitempty,iso8601" valid:"optional"`
	Metadata   []FieldMetaData      `json:"metadata,omitempty" jsonapi:"attr,metadata,omitempty" valid:"optional"` // a JSON field containing information about data field origin and update time
	Position   CommonGeoJSONPoint   `json:"position,omitempty" jsonapi:"attr,position,omitempty" valid:"optional"`
	UpdatedAt  time.Time            `json:"updatedAt,omitempty" jsonapi:"attr,updatedAt,omitempty,iso8601" valid:"optional"`
}

// POIType POI type this applies to
type POIType string

// POIs ...
type POIs []*POI

// Policies ...
type Policies []*Policy

// Policy ...
type Policy struct {
	ID        string          `jsonapi:"primary,policies,omitempty" valid:"uuid,optional"` // Policy ID
	CountryID CommonCountryID `json:"countryId,omitempty" jsonapi:"attr,countryId,omitempty" valid:"optional"`
	CreatedAt time.Time       `json:"createdAt,omitempty" jsonapi:"attr,createdAt,omitempty,iso8601" valid:"optional"` // Time of POI creation in (iso8601 without zone - expects UTC)
	PoiType   POIType         `json:"poiType,omitempty" jsonapi:"attr,poiType,omitempty" valid:"optional"`
	Rules     []PolicyRule    `json:"rules,omitempty" jsonapi:"attr,rules,omitempty" valid:"optional"`
	UserID    string          `json:"userId,omitempty" jsonapi:"attr,userId,omitempty" valid:"optional,uuid"` // Tracks who did last change
}

// PolicyRule ...
type PolicyRule struct {
	Field      FieldName            `json:"field,omitempty" jsonapi:"attr,field,omitempty" valid:"required"`
	Priorities []PolicyRulePriority `json:"priorities,omitempty" jsonapi:"attr,priorities,omitempty" valid:"required"`
}

// PolicyRulePriority ...
type PolicyRulePriority struct {
	SourceID   string  `json:"sourceId,omitempty" jsonapi:"attr,sourceId,omitempty" valid:"required,uuid"` // Tracks who did last change
	TimeToLive float64 `json:"timeToLive,omitempty" jsonapi:"attr,timeToLive,omitempty" valid:"optional"`  // Time to live in seconds (in relation to other entries)
}

// Source ...
type Source struct {
	ID         string      `jsonapi:"primary,sources,omitempty" valid:"uuid,optional"` // Source ID
	CreatedAt  time.Time   `json:"createdAt,omitempty" jsonapi:"attr,createdAt,omitempty,iso8601" valid:"optional"`
	LastDataAt time.Time   `json:"lastDataAt,omitempty" jsonapi:"attr,lastDataAt,omitempty,iso8601" valid:"optional"` // timestamp of last import from source
	Name       string      `json:"name,omitempty" jsonapi:"attr,name,omitempty" valid:"optional"`                     // source name, unique
	PoiType    POIType     `json:"poiType,omitempty" jsonapi:"attr,poiType,omitempty" valid:"optional"`
	Schema     []FieldName `json:"schema,omitempty" jsonapi:"attr,schema,omitempty" valid:"optional"` // JSON field describing the structure of the updates sent by the data source
	UpdatedAt  time.Time   `json:"updatedAt,omitempty" jsonapi:"attr,updatedAt,omitempty,iso8601" valid:"optional"`
}

// Sources ...
type Sources []*Source

// Subscription ...
type Subscription struct {
	ID        string  `jsonapi:"primary,subscription,omitempty" valid:"uuid,optional"`                 // POI Subscription ID
	PushToken string  `json:"pushToken,omitempty" jsonapi:"attr,pushToken,omitempty" valid:"optional"` // Firebase registration token
	Ttl       float64 `json:"ttl,omitempty" jsonapi:"attr,ttl,omitempty" valid:"optional"`             // TTL value for the subscription in minutes
}

/*
SubscriptionRequestArea Once entered, a notification is sent
*/
type SubscriptionRequestArea struct {
	Coordinates [][]float32 `json:"coordinates,omitempty" jsonapi:"attr,coordinates,omitempty" valid:"required"` /*
	Polygon coordinates with 4 or more positions. The first and last positions are equivalent (they represent equivalent points)
	*/
	Type string `json:"type,omitempty" jsonapi:"attr,type,omitempty" valid:"required,in(Polygon)"`
}

// SubscriptionRequest ...
type SubscriptionRequest struct {
	ID   string                  `jsonapi:"primary,subscription,omitempty" valid:"uuid,optional"`       // Example: "0c5b01d8-8dde-4d9f-be20-0865766bae6e"
	Area SubscriptionRequestArea `json:"area,omitempty" jsonapi:"attr,area,omitempty" valid:"required"` /*
	Once entered, a notification is sent
	*/
	PushToken string   `json:"pushToken,omitempty" jsonapi:"attr,pushToken,omitempty" valid:"required"`                                  // Firebase registration token
	Types     []string `json:"types,omitempty" jsonapi:"attr,types,omitempty" valid:"required,in(gasStation|movableCamera|fixedCamera)"` /*
	Filter for POI types contained in the push notification. An empty array indicates, that all POI types are allowed
	*/
}

// Currency ...
type Currency string

// FuelAmountUnit ...
type FuelAmountUnit string

/*
GetAppsHandler handles request/response marshaling and validation for
 Get /beta/apps
*/
func GetAppsHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("GetAppsHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "GetAppsHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := getAppsResponseWriter{
			ResponseWriter: metrics.NewMetric("poi", "/beta/apps", w, r),
		}
		request := GetAppsRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamPageNumber,
			Location: runtime.ScanInQuery,
			Input:    vars["page[number]"],
			Name:     "page[number]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamPageSize,
			Location: runtime.ScanInQuery,
			Input:    vars["page[size]"],
			Name:     "page[size]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterAppType,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[appType]"],
			Name:     "filter[appType]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterQuery,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[query]"],
			Name:     "filter[query]",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Invoke service that implements the business logic
		err := service.GetApps(ctx, &writer, &request)
		if err != nil {
			errors.HandleError(err, "GetAppsHandler", w, r)
		}
	})
}

/*
CreateAppHandler handles request/response marshaling and validation for
 Post /beta/apps
*/
func CreateAppHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("CreateAppHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "CreateAppHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := createAppResponseWriter{
			ResponseWriter: metrics.NewMetric("poi", "/beta/apps", w, r),
		}
		request := CreateAppRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Unmarshal the service request body
		if runtime.Unmarshal(w, r, &request.Content) {
			// Invoke service that implements the business logic
			err := service.CreateApp(ctx, &writer, &request)
			if err != nil {
				errors.HandleError(err, "CreateAppHandler", w, r)
			}
		}
	})
}

/*
CheckForPaceAppHandler handles request/response marshaling and validation for
 Get /beta/apps/query
*/
func CheckForPaceAppHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("CheckForPaceAppHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "CheckForPaceAppHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := checkForPaceAppResponseWriter{
			ResponseWriter: metrics.NewMetric("poi", "/beta/apps/query", w, r),
		}
		request := CheckForPaceAppRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamPageNumber,
			Location: runtime.ScanInQuery,
			Input:    vars["page[number]"],
			Name:     "page[number]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamPageSize,
			Location: runtime.ScanInQuery,
			Input:    vars["page[size]"],
			Name:     "page[size]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterLatitude,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[latitude]"],
			Name:     "filter[latitude]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterLongitude,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[longitude]"],
			Name:     "filter[longitude]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterGpsSource,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[gpsSource]"],
			Name:     "filter[gpsSource]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterAppType,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[appType]"],
			Name:     "filter[appType]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterAccuracy,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[accuracy]"],
			Name:     "filter[accuracy]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterDeviation,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[deviation]"],
			Name:     "filter[deviation]",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Invoke service that implements the business logic
		err := service.CheckForPaceApp(ctx, &writer, &request)
		if err != nil {
			errors.HandleError(err, "CheckForPaceAppHandler", w, r)
		}
	})
}

/*
DeleteAppHandler handles request/response marshaling and validation for
 Delete /beta/apps/{appID}
*/
func DeleteAppHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("DeleteAppHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "DeleteAppHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := deleteAppResponseWriter{
			ResponseWriter: metrics.NewMetric("poi", "/beta/apps/{appID}", w, r),
		}
		request := DeleteAppRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamAppID,
			Location: runtime.ScanInPath,
			Input:    vars["appID"],
			Name:     "appID",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Invoke service that implements the business logic
		err := service.DeleteApp(ctx, &writer, &request)
		if err != nil {
			errors.HandleError(err, "DeleteAppHandler", w, r)
		}
	})
}

/*
GetAppHandler handles request/response marshaling and validation for
 Get /beta/apps/{appID}
*/
func GetAppHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("GetAppHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "GetAppHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := getAppResponseWriter{
			ResponseWriter: metrics.NewMetric("poi", "/beta/apps/{appID}", w, r),
		}
		request := GetAppRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamAppID,
			Location: runtime.ScanInPath,
			Input:    vars["appID"],
			Name:     "appID",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Invoke service that implements the business logic
		err := service.GetApp(ctx, &writer, &request)
		if err != nil {
			errors.HandleError(err, "GetAppHandler", w, r)
		}
	})
}

/*
UpdateAppHandler handles request/response marshaling and validation for
 Put /beta/apps/{appID}
*/
func UpdateAppHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("UpdateAppHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "UpdateAppHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := updateAppResponseWriter{
			ResponseWriter: metrics.NewMetric("poi", "/beta/apps/{appID}", w, r),
		}
		request := UpdateAppRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamAppID,
			Location: runtime.ScanInPath,
			Input:    vars["appID"],
			Name:     "appID",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Unmarshal the service request body
		if runtime.Unmarshal(w, r, &request.Content) {
			// Invoke service that implements the business logic
			err := service.UpdateApp(ctx, &writer, &request)
			if err != nil {
				errors.HandleError(err, "UpdateAppHandler", w, r)
			}
		}
	})
}

/*
GetAppPOIsRelationshipsHandler handles request/response marshaling and validation for
 Get /beta/apps/{appID}/relationships/pois
*/
func GetAppPOIsRelationshipsHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("GetAppPOIsRelationshipsHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "GetAppPOIsRelationshipsHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := getAppPOIsRelationshipsResponseWriter{
			ResponseWriter: metrics.NewMetric("poi", "/beta/apps/{appID}/relationships/pois", w, r),
		}
		request := GetAppPOIsRelationshipsRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamAppID,
			Location: runtime.ScanInPath,
			Input:    vars["appID"],
			Name:     "appID",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Invoke service that implements the business logic
		err := service.GetAppPOIsRelationships(ctx, &writer, &request)
		if err != nil {
			errors.HandleError(err, "GetAppPOIsRelationshipsHandler", w, r)
		}
	})
}

/*
UpdateAppPOIsRelationshipsHandler handles request/response marshaling and validation for
 Patch /beta/apps/{appID}/relationships/pois
*/
func UpdateAppPOIsRelationshipsHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("UpdateAppPOIsRelationshipsHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "UpdateAppPOIsRelationshipsHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := updateAppPOIsRelationshipsResponseWriter{
			ResponseWriter: metrics.NewMetric("poi", "/beta/apps/{appID}/relationships/pois", w, r),
		}
		request := UpdateAppPOIsRelationshipsRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamAppID,
			Location: runtime.ScanInPath,
			Input:    vars["appID"],
			Name:     "appID",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Unmarshal the service request body
		if runtime.Unmarshal(w, r, &request.Content) {
			// Invoke service that implements the business logic
			err := service.UpdateAppPOIsRelationships(ctx, &writer, &request)
			if err != nil {
				errors.HandleError(err, "UpdateAppPOIsRelationshipsHandler", w, r)
			}
		}
	})
}

/*
GetEventsHandler handles request/response marshaling and validation for
 Get /beta/events
*/
func GetEventsHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("GetEventsHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "GetEventsHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := getEventsResponseWriter{
			ResponseWriter: metrics.NewMetric("poi", "/beta/events", w, r),
		}
		request := GetEventsRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamPageNumber,
			Location: runtime.ScanInQuery,
			Input:    vars["page[number]"],
			Name:     "page[number]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamPageSize,
			Location: runtime.ScanInQuery,
			Input:    vars["page[size]"],
			Name:     "page[size]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterSourceID,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[sourceId]"],
			Name:     "filter[sourceId]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterUserID,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[userId]"],
			Name:     "filter[userId]",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Invoke service that implements the business logic
		err := service.GetEvents(ctx, &writer, &request)
		if err != nil {
			errors.HandleError(err, "GetEventsHandler", w, r)
		}
	})
}

/*
GetGasStationsHandler handles request/response marshaling and validation for
 Get /beta/gas-stations
*/
func GetGasStationsHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("GetGasStationsHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "GetGasStationsHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := getGasStationsResponseWriter{
			ResponseWriter: metrics.NewMetric("poi", "/beta/gas-stations", w, r),
		}
		request := GetGasStationsRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamPageNumber,
			Location: runtime.ScanInQuery,
			Input:    vars["page[number]"],
			Name:     "page[number]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamPageSize,
			Location: runtime.ScanInQuery,
			Input:    vars["page[size]"],
			Name:     "page[size]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterPoiType,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[poiType]"],
			Name:     "filter[poiType]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterAppType,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[appType]"],
			Name:     "filter[appType]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterGpsSource,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[gpsSource]"],
			Name:     "filter[gpsSource]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamInclude,
			Location: runtime.ScanInQuery,
			Input:    vars["include"],
			Name:     "include",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterLatitude,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[latitude]"],
			Name:     "filter[latitude]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterLongitude,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[longitude]"],
			Name:     "filter[longitude]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterRadius,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[radius]"],
			Name:     "filter[radius]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterAccuracy,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[accuracy]"],
			Name:     "filter[accuracy]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterDeviation,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[deviation]"],
			Name:     "filter[deviation]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterBoundingBox,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[boundingBox]"],
			Name:     "filter[boundingBox]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterPath,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[path]"],
			Name:     "filter[path]",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Invoke service that implements the business logic
		err := service.GetGasStations(ctx, &writer, &request)
		if err != nil {
			errors.HandleError(err, "GetGasStationsHandler", w, r)
		}
	})
}

/*
GetGasStationHandler handles request/response marshaling and validation for
 Get /beta/gas-stations/{id}
*/
func GetGasStationHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("GetGasStationHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "GetGasStationHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := getGasStationResponseWriter{
			ResponseWriter: metrics.NewMetric("poi", "/beta/gas-stations/{id}", w, r),
		}
		request := GetGasStationRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamID,
			Location: runtime.ScanInPath,
			Input:    vars["id"],
			Name:     "id",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Invoke service that implements the business logic
		err := service.GetGasStation(ctx, &writer, &request)
		if err != nil {
			errors.HandleError(err, "GetGasStationHandler", w, r)
		}
	})
}

/*
GetPoisHandler handles request/response marshaling and validation for
 Get /beta/pois
*/
func GetPoisHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("GetPoisHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "GetPoisHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := getPoisResponseWriter{
			ResponseWriter: metrics.NewMetric("poi", "/beta/pois", w, r),
		}
		request := GetPoisRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamPageNumber,
			Location: runtime.ScanInQuery,
			Input:    vars["page[number]"],
			Name:     "page[number]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamPageSize,
			Location: runtime.ScanInQuery,
			Input:    vars["page[size]"],
			Name:     "page[size]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterPoiType,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[poiType]"],
			Name:     "filter[poiType]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterAppID,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[appId]"],
			Name:     "filter[appId]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterQuery,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[query]"],
			Name:     "filter[query]",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Invoke service that implements the business logic
		err := service.GetPois(ctx, &writer, &request)
		if err != nil {
			errors.HandleError(err, "GetPoisHandler", w, r)
		}
	})
}

/*
GetPoiHandler handles request/response marshaling and validation for
 Get /beta/pois/{poiId}
*/
func GetPoiHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("GetPoiHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "GetPoiHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := getPoiResponseWriter{
			ResponseWriter: metrics.NewMetric("poi", "/beta/pois/{poiId}", w, r),
		}
		request := GetPoiRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamPoiID,
			Location: runtime.ScanInPath,
			Input:    vars["poiId"],
			Name:     "poiId",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Invoke service that implements the business logic
		err := service.GetPoi(ctx, &writer, &request)
		if err != nil {
			errors.HandleError(err, "GetPoiHandler", w, r)
		}
	})
}

/*
ChangePoiHandler handles request/response marshaling and validation for
 Patch /beta/pois/{poiId}
*/
func ChangePoiHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("ChangePoiHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "ChangePoiHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := changePoiResponseWriter{
			ResponseWriter: metrics.NewMetric("poi", "/beta/pois/{poiId}", w, r),
		}
		request := ChangePoiRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamPoiID,
			Location: runtime.ScanInPath,
			Input:    vars["poiId"],
			Name:     "poiId",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Unmarshal the service request body
		if runtime.Unmarshal(w, r, &request.Content) {
			// Invoke service that implements the business logic
			err := service.ChangePoi(ctx, &writer, &request)
			if err != nil {
				errors.HandleError(err, "ChangePoiHandler", w, r)
			}
		}
	})
}

/*
GetPoliciesHandler handles request/response marshaling and validation for
 Get /beta/policies
*/
func GetPoliciesHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("GetPoliciesHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "GetPoliciesHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := getPoliciesResponseWriter{
			ResponseWriter: metrics.NewMetric("poi", "/beta/policies", w, r),
		}
		request := GetPoliciesRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamPageNumber,
			Location: runtime.ScanInQuery,
			Input:    vars["page[number]"],
			Name:     "page[number]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamPageSize,
			Location: runtime.ScanInQuery,
			Input:    vars["page[size]"],
			Name:     "page[size]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterPoiType,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[poiType]"],
			Name:     "filter[poiType]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterCountryID,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[countryId]"],
			Name:     "filter[countryId]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterUserID,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[userId]"],
			Name:     "filter[userId]",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Invoke service that implements the business logic
		err := service.GetPolicies(ctx, &writer, &request)
		if err != nil {
			errors.HandleError(err, "GetPoliciesHandler", w, r)
		}
	})
}

/*
CreatePolicyHandler handles request/response marshaling and validation for
 Post /beta/policies
*/
func CreatePolicyHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("CreatePolicyHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "CreatePolicyHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := createPolicyResponseWriter{
			ResponseWriter: metrics.NewMetric("poi", "/beta/policies", w, r),
		}
		request := CreatePolicyRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Unmarshal the service request body
		if runtime.Unmarshal(w, r, &request.Content) {
			// Invoke service that implements the business logic
			err := service.CreatePolicy(ctx, &writer, &request)
			if err != nil {
				errors.HandleError(err, "CreatePolicyHandler", w, r)
			}
		}
	})
}

/*
GetPolicyHandler handles request/response marshaling and validation for
 Get /beta/policies/{policyId}
*/
func GetPolicyHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("GetPolicyHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "GetPolicyHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := getPolicyResponseWriter{
			ResponseWriter: metrics.NewMetric("poi", "/beta/policies/{policyId}", w, r),
		}
		request := GetPolicyRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamPolicyID,
			Location: runtime.ScanInPath,
			Input:    vars["policyId"],
			Name:     "policyId",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Invoke service that implements the business logic
		err := service.GetPolicy(ctx, &writer, &request)
		if err != nil {
			errors.HandleError(err, "GetPolicyHandler", w, r)
		}
	})
}

/*
GetSourcesHandler handles request/response marshaling and validation for
 Get /beta/sources
*/
func GetSourcesHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("GetSourcesHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "GetSourcesHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := getSourcesResponseWriter{
			ResponseWriter: metrics.NewMetric("poi", "/beta/sources", w, r),
		}
		request := GetSourcesRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamPageNumber,
			Location: runtime.ScanInQuery,
			Input:    vars["page[number]"],
			Name:     "page[number]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamPageSize,
			Location: runtime.ScanInQuery,
			Input:    vars["page[size]"],
			Name:     "page[size]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterPoiType,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[poiType]"],
			Name:     "filter[poiType]",
		}, &runtime.ScanParameter{
			Data:     &request.ParamFilterName,
			Location: runtime.ScanInQuery,
			Input:    vars["filter[name]"],
			Name:     "filter[name]",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Invoke service that implements the business logic
		err := service.GetSources(ctx, &writer, &request)
		if err != nil {
			errors.HandleError(err, "GetSourcesHandler", w, r)
		}
	})
}

/*
CreateSourceHandler handles request/response marshaling and validation for
 Post /beta/sources
*/
func CreateSourceHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("CreateSourceHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "CreateSourceHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := createSourceResponseWriter{
			ResponseWriter: metrics.NewMetric("poi", "/beta/sources", w, r),
		}
		request := CreateSourceRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Unmarshal the service request body
		if runtime.Unmarshal(w, r, &request.Content) {
			// Invoke service that implements the business logic
			err := service.CreateSource(ctx, &writer, &request)
			if err != nil {
				errors.HandleError(err, "CreateSourceHandler", w, r)
			}
		}
	})
}

/*
DeleteSourceHandler handles request/response marshaling and validation for
 Delete /beta/sources/{sourceId}
*/
func DeleteSourceHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("DeleteSourceHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "DeleteSourceHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := deleteSourceResponseWriter{
			ResponseWriter: metrics.NewMetric("poi", "/beta/sources/{sourceId}", w, r),
		}
		request := DeleteSourceRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamSourceID,
			Location: runtime.ScanInPath,
			Input:    vars["sourceId"],
			Name:     "sourceId",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Invoke service that implements the business logic
		err := service.DeleteSource(ctx, &writer, &request)
		if err != nil {
			errors.HandleError(err, "DeleteSourceHandler", w, r)
		}
	})
}

/*
GetSourceHandler handles request/response marshaling and validation for
 Get /beta/sources/{sourceId}
*/
func GetSourceHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("GetSourceHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "GetSourceHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := getSourceResponseWriter{
			ResponseWriter: metrics.NewMetric("poi", "/beta/sources/{sourceId}", w, r),
		}
		request := GetSourceRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamSourceID,
			Location: runtime.ScanInPath,
			Input:    vars["sourceId"],
			Name:     "sourceId",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Invoke service that implements the business logic
		err := service.GetSource(ctx, &writer, &request)
		if err != nil {
			errors.HandleError(err, "GetSourceHandler", w, r)
		}
	})
}

/*
UpdateSourceHandler handles request/response marshaling and validation for
 Put /beta/sources/{sourceId}
*/
func UpdateSourceHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("UpdateSourceHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "UpdateSourceHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := updateSourceResponseWriter{
			ResponseWriter: metrics.NewMetric("poi", "/beta/sources/{sourceId}", w, r),
		}
		request := UpdateSourceRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		vars := mux.Vars(r)
		if !runtime.ScanParameters(w, r, &runtime.ScanParameter{
			Data:     &request.ParamSourceID,
			Location: runtime.ScanInPath,
			Input:    vars["sourceId"],
			Name:     "sourceId",
		}) {
			return
		}
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Unmarshal the service request body
		if runtime.Unmarshal(w, r, &request.Content) {
			// Invoke service that implements the business logic
			err := service.UpdateSource(ctx, &writer, &request)
			if err != nil {
				errors.HandleError(err, "UpdateSourceHandler", w, r)
			}
		}
	})
}

/*
CreateSubscriptionHandler handles request/response marshaling and validation for
 Post /beta/subscriptions
*/
func CreateSubscriptionHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("CreateSubscriptionHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "CreateSubscriptionHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := createSubscriptionResponseWriter{
			ResponseWriter: metrics.NewMetric("poi", "/beta/subscriptions", w, r),
		}
		request := CreateSubscriptionRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters
		if !runtime.ValidateParameters(w, r, &request) {
			return // invalid request stop further processing
		}

		// Unmarshal the service request body
		if runtime.Unmarshal(w, r, &request.Content) {
			// Invoke service that implements the business logic
			err := service.CreateSubscription(ctx, &writer, &request)
			if err != nil {
				errors.HandleError(err, "CreateSubscriptionHandler", w, r)
			}
		}
	})
}

/*
GetTilesHandler handles request/response marshaling and validation for
 Post /beta/tiles/query
*/
func GetTilesHandler(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer errors.HandleRequest("GetTilesHandler", w, r)

		// Trace the service function handler execution
		handlerSpan, ctx := opentracing.StartSpanFromContext(r.Context(), "GetTilesHandler")
		defer handlerSpan.Finish()

		// Setup context, response writer and request type
		writer := getTilesResponseWriter{
			ResponseWriter: metrics.NewMetric("poi", "/beta/tiles/query", w, r),
		}
		request := GetTilesRequest{
			Request: r.WithContext(ctx),
		}

		// Scan and validate incoming request parameters

		// Invoke service that implements the business logic
		err := service.GetTiles(ctx, &writer, &request)
		if err != nil {
			errors.HandleError(err, "GetTilesHandler", w, r)
		}
	})
}

/*
GetAppsResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetAppsResponseWriter interface {
	http.ResponseWriter
	OK(LocationBasedApps)
	BadRequest(error)
}
type getAppsResponseWriter struct {
	http.ResponseWriter
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *getAppsResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *getAppsResponseWriter) OK(data LocationBasedApps) {
	runtime.Marshal(w, data, 200)
}

/*
GetAppsRequest is a standard http.Request extended with the
un-marshaled content object
*/
type GetAppsRequest struct {
	Request            *http.Request `valid:"-"`
	ParamPageNumber    int64         `valid:"optional"`
	ParamPageSize      int64         `valid:"optional"`
	ParamFilterAppType string        `valid:"optional,in(fueling)"`
	ParamFilterQuery   string        `valid:"optional"`
}

/*
CreateAppResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type CreateAppResponseWriter interface {
	http.ResponseWriter
	OK(*LocationBasedApp)
	BadRequest(error)
}
type createAppResponseWriter struct {
	http.ResponseWriter
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *createAppResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// OK responds with jsonapi marshaled data (HTTP code 201)
func (w *createAppResponseWriter) OK(data *LocationBasedApp) {
	runtime.Marshal(w, data, 201)
}

// CreateAppRequest ...
type CreateAppRequest struct {
	Request *http.Request    `valid:"-"`
	Content LocationBasedApp `valid:"-"`
}

/*
CheckForPaceAppResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type CheckForPaceAppResponseWriter interface {
	http.ResponseWriter
	OK(LocationBasedApps)
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
func (w *checkForPaceAppResponseWriter) OK(data LocationBasedApps) {
	runtime.Marshal(w, data, 200)
}

/*
CheckForPaceAppRequest is a standard http.Request extended with the
un-marshaled content object
*/
type CheckForPaceAppRequest struct {
	Request              *http.Request `valid:"-"`
	ParamPageNumber      int64         `valid:"optional"`
	ParamPageSize        int64         `valid:"optional"`
	ParamFilterLatitude  float32       `valid:"required"`
	ParamFilterLongitude float32       `valid:"required"`
	ParamFilterGpsSource string        `valid:"required,in(raw|mapMatched)"`
	ParamFilterAppType   string        `valid:"required,in(fueling)"`
	ParamFilterAccuracy  float32       `valid:"optional"`
	ParamFilterDeviation float32       `valid:"optional"`
}

/*
DeleteAppResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type DeleteAppResponseWriter interface {
	http.ResponseWriter
	OK()
	NotFound(error)
}
type deleteAppResponseWriter struct {
	http.ResponseWriter
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *deleteAppResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// OK responds with empty response (HTTP code 204)
func (w *deleteAppResponseWriter) OK() {
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(204)
}

/*
DeleteAppRequest is a standard http.Request extended with the
un-marshaled content object
*/
type DeleteAppRequest struct {
	Request    *http.Request `valid:"-"`
	ParamAppID string        `valid:"optional,uuid"`
}

/*
GetAppResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetAppResponseWriter interface {
	http.ResponseWriter
	OK(*LocationBasedApp)
	BadRequest(error)
	NotFound(error)
}
type getAppResponseWriter struct {
	http.ResponseWriter
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *getAppResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *getAppResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *getAppResponseWriter) OK(data *LocationBasedApp) {
	runtime.Marshal(w, data, 200)
}

/*
GetAppRequest is a standard http.Request extended with the
un-marshaled content object
*/
type GetAppRequest struct {
	Request    *http.Request `valid:"-"`
	ParamAppID string        `valid:"optional,uuid"`
}

/*
UpdateAppResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type UpdateAppResponseWriter interface {
	http.ResponseWriter
	OK(*LocationBasedApp)
	BadRequest(error)
	NotFound(error)
}
type updateAppResponseWriter struct {
	http.ResponseWriter
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *updateAppResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *updateAppResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *updateAppResponseWriter) OK(data *LocationBasedApp) {
	runtime.Marshal(w, data, 200)
}

// UpdateAppRequest ...
type UpdateAppRequest struct {
	Request    *http.Request    `valid:"-"`
	Content    LocationBasedApp `valid:"-"`
	ParamAppID string           `valid:"optional,uuid"`
}

/*
GetAppPOIsRelationshipsResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetAppPOIsRelationshipsResponseWriter interface {
	http.ResponseWriter
	OK(AppPOIsRelationships)
	BadRequest(error)
}
type getAppPOIsRelationshipsResponseWriter struct {
	http.ResponseWriter
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *getAppPOIsRelationshipsResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *getAppPOIsRelationshipsResponseWriter) OK(data AppPOIsRelationships) {
	runtime.Marshal(w, data, 200)
}

/*
GetAppPOIsRelationshipsRequest is a standard http.Request extended with the
un-marshaled content object
*/
type GetAppPOIsRelationshipsRequest struct {
	Request    *http.Request `valid:"-"`
	ParamAppID string        `valid:"optional,uuid"`
}

/*
UpdateAppPOIsRelationshipsResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type UpdateAppPOIsRelationshipsResponseWriter interface {
	http.ResponseWriter
	OK(AppPOIsRelationships)
	BadRequest(error)
	NotFound(error)
}
type updateAppPOIsRelationshipsResponseWriter struct {
	http.ResponseWriter
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *updateAppPOIsRelationshipsResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *updateAppPOIsRelationshipsResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *updateAppPOIsRelationshipsResponseWriter) OK(data AppPOIsRelationships) {
	runtime.Marshal(w, data, 200)
}

// UpdateAppPOIsRelationshipsRequest ...
type UpdateAppPOIsRelationshipsRequest struct {
	Request    *http.Request        `valid:"-"`
	Content    AppPOIsRelationships `valid:"-"`
	ParamAppID string               `valid:"optional,uuid"`
}

/*
GetEventsResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetEventsResponseWriter interface {
	http.ResponseWriter
	OK(Events)
}
type getEventsResponseWriter struct {
	http.ResponseWriter
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *getEventsResponseWriter) OK(data Events) {
	runtime.Marshal(w, data, 200)
}

/*
GetEventsRequest is a standard http.Request extended with the
un-marshaled content object
*/
type GetEventsRequest struct {
	Request             *http.Request `valid:"-"`
	ParamPageNumber     int64         `valid:"optional"`
	ParamPageSize       int64         `valid:"optional"`
	ParamFilterSourceID string        `valid:"optional,uuid"`
	ParamFilterUserID   string        `valid:"optional,uuid"`
}

/*
GetGasStationsResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetGasStationsResponseWriter interface {
	http.ResponseWriter
	OK(GasStations)
	BadRequest(error)
}
type getGasStationsResponseWriter struct {
	http.ResponseWriter
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *getGasStationsResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *getGasStationsResponseWriter) OK(data GasStations) {
	runtime.Marshal(w, data, 200)
}

/*
GetGasStationsRequest is a standard http.Request extended with the
un-marshaled content object
*/
type GetGasStationsRequest struct {
	Request                *http.Request `valid:"-"`
	ParamPageNumber        int64         `valid:"optional"`
	ParamPageSize          int64         `valid:"optional"`
	ParamFilterPoiType     string        `valid:"required,in(gasStation)"`
	ParamFilterAppType     []string      `valid:"required,in(fueling)"`
	ParamFilterGpsSource   string        `valid:"required,in(raw|mapMatched)"`
	ParamInclude           string        `valid:"required,in(insideAppArea)"`
	ParamFilterLatitude    float32       `valid:"optional"`
	ParamFilterLongitude   float32       `valid:"optional"`
	ParamFilterRadius      float32       `valid:"optional"`
	ParamFilterAccuracy    float32       `valid:"optional"`
	ParamFilterDeviation   float32       `valid:"optional"`
	ParamFilterBoundingBox []float32     `valid:"optional"`
	ParamFilterPath        string        `valid:"optional"`
}

/*
GetGasStationResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetGasStationResponseWriter interface {
	http.ResponseWriter
	OK(*GasStation)
	NotFound(error)
}
type getGasStationResponseWriter struct {
	http.ResponseWriter
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *getGasStationResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *getGasStationResponseWriter) OK(data *GasStation) {
	runtime.Marshal(w, data, 200)
}

/*
GetGasStationRequest is a standard http.Request extended with the
un-marshaled content object
*/
type GetGasStationRequest struct {
	Request *http.Request `valid:"-"`
	ParamID string        `valid:"required,uuid"`
}

/*
GetPoisResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetPoisResponseWriter interface {
	http.ResponseWriter
	OK(POIs)
	BadRequest(error)
}
type getPoisResponseWriter struct {
	http.ResponseWriter
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *getPoisResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *getPoisResponseWriter) OK(data POIs) {
	runtime.Marshal(w, data, 200)
}

/*
GetPoisRequest is a standard http.Request extended with the
un-marshaled content object
*/
type GetPoisRequest struct {
	Request            *http.Request `valid:"-"`
	ParamPageNumber    int64         `valid:"optional"`
	ParamPageSize      int64         `valid:"optional"`
	ParamFilterPoiType POIType       `valid:"optional"`
	ParamFilterAppID   string        `valid:"optional,uuid"`
	ParamFilterQuery   string        `valid:"optional"`
}

/*
GetPoiResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetPoiResponseWriter interface {
	http.ResponseWriter
	OK(*POI)
	BadRequest(error)
	NotFound(error)
}
type getPoiResponseWriter struct {
	http.ResponseWriter
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *getPoiResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *getPoiResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *getPoiResponseWriter) OK(data *POI) {
	runtime.Marshal(w, data, 200)
}

/*
GetPoiRequest is a standard http.Request extended with the
un-marshaled content object
*/
type GetPoiRequest struct {
	Request    *http.Request `valid:"-"`
	ParamPoiID string        `valid:"optional,uuid"`
}

/*
ChangePoiResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type ChangePoiResponseWriter interface {
	http.ResponseWriter
	OK(*POI)
	BadRequest(error)
	NotFound(error)
}
type changePoiResponseWriter struct {
	http.ResponseWriter
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *changePoiResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *changePoiResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *changePoiResponseWriter) OK(data *POI) {
	runtime.Marshal(w, data, 200)
}

// ChangePoiRequest ...
type ChangePoiRequest struct {
	Request    *http.Request `valid:"-"`
	Content    POI           `valid:"-"`
	ParamPoiID string        `valid:"optional,uuid"`
}

/*
GetPoliciesResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetPoliciesResponseWriter interface {
	http.ResponseWriter
	OK(Policies)
	BadRequest(error)
}
type getPoliciesResponseWriter struct {
	http.ResponseWriter
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *getPoliciesResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *getPoliciesResponseWriter) OK(data Policies) {
	runtime.Marshal(w, data, 200)
}

/*
GetPoliciesRequest is a standard http.Request extended with the
un-marshaled content object
*/
type GetPoliciesRequest struct {
	Request              *http.Request `valid:"-"`
	ParamPageNumber      int64         `valid:"optional"`
	ParamPageSize        int64         `valid:"optional"`
	ParamFilterPoiType   POIType       `valid:"optional"`
	ParamFilterCountryID string        `valid:"optional"`
	ParamFilterUserID    string        `valid:"optional,uuid"`
}

/*
CreatePolicyResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type CreatePolicyResponseWriter interface {
	http.ResponseWriter
	OK(*Policy)
	BadRequest(error)
}
type createPolicyResponseWriter struct {
	http.ResponseWriter
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *createPolicyResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// OK responds with jsonapi marshaled data (HTTP code 201)
func (w *createPolicyResponseWriter) OK(data *Policy) {
	runtime.Marshal(w, data, 201)
}

// CreatePolicyRequest ...
type CreatePolicyRequest struct {
	Request *http.Request `valid:"-"`
	Content Policy        `valid:"-"`
}

/*
GetPolicyResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetPolicyResponseWriter interface {
	http.ResponseWriter
	OK(*Policy)
	BadRequest(error)
	NotFound(error)
}
type getPolicyResponseWriter struct {
	http.ResponseWriter
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *getPolicyResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *getPolicyResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *getPolicyResponseWriter) OK(data *Policy) {
	runtime.Marshal(w, data, 200)
}

/*
GetPolicyRequest is a standard http.Request extended with the
un-marshaled content object
*/
type GetPolicyRequest struct {
	Request       *http.Request `valid:"-"`
	ParamPolicyID string        `valid:"optional,uuid"`
}

/*
GetSourcesResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetSourcesResponseWriter interface {
	http.ResponseWriter
	OK(Sources)
	BadRequest(error)
}
type getSourcesResponseWriter struct {
	http.ResponseWriter
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *getSourcesResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *getSourcesResponseWriter) OK(data Sources) {
	runtime.Marshal(w, data, 200)
}

/*
GetSourcesRequest is a standard http.Request extended with the
un-marshaled content object
*/
type GetSourcesRequest struct {
	Request            *http.Request `valid:"-"`
	ParamPageNumber    int64         `valid:"optional"`
	ParamPageSize      int64         `valid:"optional"`
	ParamFilterPoiType POIType       `valid:"optional"`
	ParamFilterName    string        `valid:"optional"`
}

/*
CreateSourceResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type CreateSourceResponseWriter interface {
	http.ResponseWriter
	OK(*Source)
	BadRequest(error)
}
type createSourceResponseWriter struct {
	http.ResponseWriter
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *createSourceResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// OK responds with jsonapi marshaled data (HTTP code 201)
func (w *createSourceResponseWriter) OK(data *Source) {
	runtime.Marshal(w, data, 201)
}

// CreateSourceRequest ...
type CreateSourceRequest struct {
	Request *http.Request `valid:"-"`
	Content Source        `valid:"-"`
}

/*
DeleteSourceResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type DeleteSourceResponseWriter interface {
	http.ResponseWriter
	OK()
	NotFound(error)
}
type deleteSourceResponseWriter struct {
	http.ResponseWriter
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *deleteSourceResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// OK responds with empty response (HTTP code 204)
func (w *deleteSourceResponseWriter) OK() {
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(204)
}

/*
DeleteSourceRequest is a standard http.Request extended with the
un-marshaled content object
*/
type DeleteSourceRequest struct {
	Request       *http.Request `valid:"-"`
	ParamSourceID string        `valid:"optional,uuid"`
}

/*
GetSourceResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetSourceResponseWriter interface {
	http.ResponseWriter
	OK(*Source)
	BadRequest(error)
	NotFound(error)
}
type getSourceResponseWriter struct {
	http.ResponseWriter
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *getSourceResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *getSourceResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *getSourceResponseWriter) OK(data *Source) {
	runtime.Marshal(w, data, 200)
}

/*
GetSourceRequest is a standard http.Request extended with the
un-marshaled content object
*/
type GetSourceRequest struct {
	Request       *http.Request `valid:"-"`
	ParamSourceID string        `valid:"optional,uuid"`
}

/*
UpdateSourceResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type UpdateSourceResponseWriter interface {
	http.ResponseWriter
	OK(*Source)
	BadRequest(error)
	NotFound(error)
}
type updateSourceResponseWriter struct {
	http.ResponseWriter
}

// NotFound responds with jsonapi error (HTTP code 404)
func (w *updateSourceResponseWriter) NotFound(err error) {
	runtime.WriteError(w, 404, err)
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *updateSourceResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// OK responds with jsonapi marshaled data (HTTP code 200)
func (w *updateSourceResponseWriter) OK(data *Source) {
	runtime.Marshal(w, data, 200)
}

// UpdateSourceRequest ...
type UpdateSourceRequest struct {
	Request       *http.Request `valid:"-"`
	Content       Source        `valid:"-"`
	ParamSourceID string        `valid:"optional,uuid"`
}

/*
CreateSubscriptionResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type CreateSubscriptionResponseWriter interface {
	http.ResponseWriter
	Created(*Subscription)
	BadRequest(error)
}
type createSubscriptionResponseWriter struct {
	http.ResponseWriter
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *createSubscriptionResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// Created responds with jsonapi marshaled data (HTTP code 201)
func (w *createSubscriptionResponseWriter) Created(data *Subscription) {
	runtime.Marshal(w, data, 201)
}

// CreateSubscriptionRequest ...
type CreateSubscriptionRequest struct {
	Request *http.Request       `valid:"-"`
	Content SubscriptionRequest `valid:"-"`
}

/*
GetTilesResponseWriter is a standard http.ResponseWriter extended with methods
to generate the respective responses easily
*/
type GetTilesResponseWriter interface {
	http.ResponseWriter
	OK()
	BadRequest(error)
}
type getTilesResponseWriter struct {
	http.ResponseWriter
}

// BadRequest responds with jsonapi error (HTTP code 400)
func (w *getTilesResponseWriter) BadRequest(err error) {
	runtime.WriteError(w, 400, err)
}

// OK responds with empty response (HTTP code 200)
func (w *getTilesResponseWriter) OK() {
	w.Header().Set("Content-Type", "application/protobuf")
	w.WriteHeader(200)
}

// GetTilesRequest ...
type GetTilesRequest struct {
	Request *http.Request `valid:"-"`
}

// Service interface for all handlers
type Service interface {
	/*
	   GetApps Returns a paginated list of apps

	   Returns a paginated list of apps optionally filtered by type and/or query
	*/
	GetApps(context.Context, GetAppsResponseWriter, *GetAppsRequest) error
	/*
	   CreateApp Creates a new application

	   Creates a new application
	*/
	CreateApp(context.Context, CreateAppResponseWriter, *CreateAppRequest) error
	/*
	   CheckForPaceApp Query for location-based apps


	   These location-based PACE apps deliver additional services for PACE customers based on their current position.
	   You can (or should) trigger this whenever:
	   * A longer stand-still is detected
	   * The engine is turned off
	   * Every 5 seconds if the user "left the road"

	   Please note that calling this API is very cheap and can be done regularly.
	*/
	CheckForPaceApp(context.Context, CheckForPaceAppResponseWriter, *CheckForPaceAppRequest) error
	/*
	   DeleteApp Deletes App with specified id

	   Deletes App with specified id
	*/
	DeleteApp(context.Context, DeleteAppResponseWriter, *DeleteAppRequest) error
	/*
	   GetApp Returns App with specified id

	   Returns App with specified id
	*/
	GetApp(context.Context, GetAppResponseWriter, *GetAppRequest) error
	/*
	   UpdateApp Updates App with specified id

	   Updates App with specified id
	*/
	UpdateApp(context.Context, UpdateAppResponseWriter, *UpdateAppRequest) error
	/*
	   GetAppPOIsRelationships Returns all POI relations for specified app id

	   Returns all POI relations for specified app id
	*/
	GetAppPOIsRelationships(context.Context, GetAppPOIsRelationshipsResponseWriter, *GetAppPOIsRelationshipsRequest) error
	/*
	   UpdateAppPOIsRelationships Update all POI relations for specified app id

	   Update all POI relations for specified app id
	*/
	UpdateAppPOIsRelationships(context.Context, UpdateAppPOIsRelationshipsResponseWriter, *UpdateAppPOIsRelationshipsRequest) error
	/*
	   GetEvents Returns a list of events

	   Returns a list of eventsoptionally filtered by poi type and/or country id and/or user id
	*/
	GetEvents(context.Context, GetEventsResponseWriter, *GetEventsRequest) error
	/*
	   GetGasStations Query for gas stations

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
	GetGasStations(context.Context, GetGasStationsResponseWriter, *GetGasStationsRequest) error
	/*
	   GetGasStation Get a specific gas station

	   Returns an individual gas station
	*/
	GetGasStation(context.Context, GetGasStationResponseWriter, *GetGasStationRequest) error
	/*
	   GetPois Returns a paginated list of POIs

	   Returns a paginated list of POIs optionally filtered by type, appId and/or query
	*/
	GetPois(context.Context, GetPoisResponseWriter, *GetPoisRequest) error
	/*
	   GetPoi Returns POI with specified id

	   Returns POI with specified id
	*/
	GetPoi(context.Context, GetPoiResponseWriter, *GetPoiRequest) error
	/*
	   ChangePoi Updates POI with specified id (only passed attributes will be updated)

	   Returns POI with specified id (only passed attributes will be updated)
	*/
	ChangePoi(context.Context, ChangePoiResponseWriter, *ChangePoiRequest) error
	/*
	   GetPolicies Returns a paginated list of policies

	   Returns a paginated list of policies optionally filtered by poi type and/or country id and/or user id
	*/
	GetPolicies(context.Context, GetPoliciesResponseWriter, *GetPoliciesRequest) error
	/*
	   CreatePolicy Creates a new policy

	   Creates a new policy
	*/
	CreatePolicy(context.Context, CreatePolicyResponseWriter, *CreatePolicyRequest) error
	/*
	   GetPolicy Returns policy with specified id

	   Returns policy with specified id
	*/
	GetPolicy(context.Context, GetPolicyResponseWriter, *GetPolicyRequest) error
	/*
	   GetSources Returns a paginated list of sources

	   Returns a paginated list of sources optionally filtered by poi type and/or name
	*/
	GetSources(context.Context, GetSourcesResponseWriter, *GetSourcesRequest) error
	/*
	   CreateSource Creates a new source

	   Creates a new source
	*/
	CreateSource(context.Context, CreateSourceResponseWriter, *CreateSourceRequest) error
	/*
	   DeleteSource Deletes source with specified id

	   Deletes source with specified id
	*/
	DeleteSource(context.Context, DeleteSourceResponseWriter, *DeleteSourceRequest) error
	/*
	   GetSource Returns source with specified id

	   Returns source with specified id
	*/
	GetSource(context.Context, GetSourceResponseWriter, *GetSourceRequest) error
	/*
	   UpdateSource Updates source with specified id

	   Updates source with specified id
	*/
	UpdateSource(context.Context, UpdateSourceResponseWriter, *UpdateSourceRequest) error
	/*
	   CreateSubscription Create a POI subscription


	   Create a POI subscription to send a push notification to the device with the specified `pushToken` once you enter the specified `area`, which is currently described by a polygon. If you specify `types`, you only get one of those POI types in the push notification. The notification contains (max 4kb)

	   ```
	   {
	     "target": "mapkit"
	     "poi": {
	       "id": "B064797C-C644-4D48-8DDD-E2D6A7D86770", # poi ID
	       "type": "movableCamera",
	       "attributes": {
	         "coordinates": [101.0, 0.0], # lat, long
	         # ... potentially more data
	       }
	     }
	   } ```
	*/
	CreateSubscription(context.Context, CreateSubscriptionResponseWriter, *CreateSubscriptionRequest) error
	/*
	   GetTiles Query for tiles


	   Get a list of map tiles in the Protobuf binary wire format.
	*/
	GetTiles(context.Context, GetTilesResponseWriter, *GetTilesRequest) error
}

/*
Router implements: PACE POI API

POI API
*/
func Router(service Service) *mux.Router {
	router := mux.NewRouter()
	// Subrouter s1 - Path: /poi
	s1 := router.PathPrefix("/poi").Subrouter()
	s1.Methods("GET").Path("/beta/apps/query").Handler(CheckForPaceAppHandler(service)).Name("CheckForPaceApp")
	s1.Methods("POST").Path("/beta/tiles/query").Handler(GetTilesHandler(service)).Name("GetTiles")
	s1.Methods("GET").Path("/beta/pois").Handler(GetPoisHandler(service)).Name("GetPois")
	s1.Methods("POST").Path("/beta/apps").Handler(CreateAppHandler(service)).Name("CreateApp")
	s1.Methods("POST").Path("/beta/subscriptions").Handler(CreateSubscriptionHandler(service)).Name("CreateSubscription")
	s1.Methods("GET").Path("/beta/sources").Handler(GetSourcesHandler(service)).Name("GetSources")
	s1.Methods("POST").Path("/beta/policies").Handler(CreatePolicyHandler(service)).Name("CreatePolicy")
	s1.Methods("GET").Path("/beta/policies").Handler(GetPoliciesHandler(service)).Name("GetPolicies")
	s1.Methods("GET").Path("/beta/events").Handler(GetEventsHandler(service)).Name("GetEvents")
	s1.Methods("GET").Path("/beta/gas-stations").Handler(GetGasStationsHandler(service)).Name("GetGasStations")
	s1.Methods("POST").Path("/beta/sources").Handler(CreateSourceHandler(service)).Name("CreateSource")
	s1.Methods("GET").Path("/beta/apps").Handler(GetAppsHandler(service)).Name("GetApps")
	s1.Methods("PATCH").Path("/beta/apps/{appID}/relationships/pois").Handler(UpdateAppPOIsRelationshipsHandler(service)).Name("UpdateAppPOIsRelationships")
	s1.Methods("GET").Path("/beta/apps/{appID}/relationships/pois").Handler(GetAppPOIsRelationshipsHandler(service)).Name("GetAppPOIsRelationships")
	s1.Methods("GET").Path("/beta/gas-stations/{id}").Handler(GetGasStationHandler(service)).Name("GetGasStation")
	s1.Methods("PATCH").Path("/beta/pois/{poiId}").Handler(ChangePoiHandler(service)).Name("ChangePoi")
	s1.Methods("GET").Path("/beta/policies/{policyId}").Handler(GetPolicyHandler(service)).Name("GetPolicy")
	s1.Methods("PUT").Path("/beta/apps/{appID}").Handler(UpdateAppHandler(service)).Name("UpdateApp")
	s1.Methods("GET").Path("/beta/apps/{appID}").Handler(GetAppHandler(service)).Name("GetApp")
	s1.Methods("DELETE").Path("/beta/sources/{sourceId}").Handler(DeleteSourceHandler(service)).Name("DeleteSource")
	s1.Methods("GET").Path("/beta/sources/{sourceId}").Handler(GetSourceHandler(service)).Name("GetSource")
	s1.Methods("PUT").Path("/beta/sources/{sourceId}").Handler(UpdateSourceHandler(service)).Name("UpdateSource")
	s1.Methods("DELETE").Path("/beta/apps/{appID}").Handler(DeleteAppHandler(service)).Name("DeleteApp")
	s1.Methods("GET").Path("/beta/pois/{poiId}").Handler(GetPoiHandler(service)).Name("GetPoi")
	return router
}
