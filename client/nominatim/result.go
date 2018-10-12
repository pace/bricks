// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/10/12 by Vincent Landgraf

package nominatim

import "fmt"

// Result contains additional information and the address of the given lat/lon pair
type Result struct {
	Error       string   `json:"error"`        // "Unable to geocode"
	PlaceID     string   `json:"place_id"`     // "84948520"
	Licence     string   `json:"licence"`      // "Data © OpenStreetMap contributors, ODbL 1.0. http://www.openstreetmap.org/copyright"
	OsmType     string   `json:"osm_type"`     // "way"
	OsmID       string   `json:"osm_id"`       // "109123011"
	Lat         string   `json:"lat"`          // "49.0081562"
	Lon         string   `json:"lon"`          // "8.39770823450571"
	PlaceRank   string   `json:"place_rank"`   // "30"
	Category    string   `json:"category"`     // "building"
	Type        string   `json:"type"`         // "yes"
	Importance  string   `json:"importance"`   // "0"
	AddressType string   `json:"addresstype"`  // "building"
	DisplayName string   `json:"display_name"` // "31, Herrenstraße, Innenstadt-West Östlicher Teil, Innenstadt-West, Karlsruhe, Regierungsbezirk Karlsruhe, Baden-Württemberg, 76133, Deutschland"
	Name        string   `json:"name"`         // null
	Address     *Address `json:"address"`
}

// Address contains parts of an address
type Address struct {
	HouseNumber   string `json:"house_number"`   //  "31"
	Road          string `json:"road"`           //  "Herrenstraße"
	Neighborhood  string `json:"neighbourhood"`  //  "Innenstadt-West Östlicher Teil"
	Suburb        string `json:"suburb"`         //  "Innenstadt-West"
	Hamlet        string `json:"hamlet"`         //  "Inderingen"
	City          string `json:"city"`           //  "Karlsruhe"
	CityDistrict  string `json:"city_district"`  //  "Frohnstetten"
	StateDistrict string `json:"state_district"` //  "Regierungsbezirk Karlsruhe"
	State         string `json:"state"`          //  "Baden-Württemberg"
	Postcode      string `json:"postcode"`       //  "76133"
	County        string `json:"county"`         //  "Kreis Karlsruhe"
	Country       string `json:"country"`        //  "Deutschland"
	CountryCode   string `json:"country_code"`   //  "de"
	Town          string `json:"town"`           //  "Sprintfield"
	Village       string `json:"village"`        //  "Stetten am kalten Markt"
}

// GermanShort returns the address in german format
func (a Address) GermanShort() string {
	return fmt.Sprintf("%s %s, %s %s", a.Road, a.HouseNumber, a.Postcode, a.CityEquivalent())
}

// CityEquivalent returns the first filled field of either 'city', 'town', 'village',
// 'hamlet' or 'suburb'.
// Based on based on https://github.com/openstreetmap/Nominatim/issues/885.
func (a Address) CityEquivalent() string {
	if a.City != "" {
		return a.City
	}

	if a.Town != "" {
		return a.Town
	}

	if a.Village != "" {
		return a.Village
	}

	if a.Hamlet != "" {
		return a.Hamlet
	}

	return a.Suburb
}
