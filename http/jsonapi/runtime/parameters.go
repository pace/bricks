// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/24 by Vincent Landgraf

package runtime

import (
	"fmt"
	"net/http"
	"strings"
)

// ScanIn help to avoid missuse using iota for the possible values
type ScanIn int

const (
	// ScanInPath hints the scanner to scan the input
	ScanInPath ScanIn = iota
	// ScanInQuery hints the scanner to scan the request url query
	ScanInQuery
)

// ScanParameter configured the ScanParameters function
type ScanParameter struct {
	// Data contains the reference to the parameter, that should
	// be scanned to
	Data interface{}
	// Where the data can be found for scanning
	Location ScanIn
	// Input must contain the value data if location is in ScanInPath
	// otherwise input contains the name of the query variable
	Input string
}

// ScanParameters scans the request using the given path parameter objects
// in case an error is encountered a 400 along with a jsonapi errors object
// is send to the ResponseWriter and false is returned. Returns true if all
// values were scanned successfully. The used scanning function is fmt.Sscan
func ScanParameters(w http.ResponseWriter, r *http.Request, parameters ...*ScanParameter) bool {
	for _, param := range parameters {
		var scanData string

		switch param.Location {
		case ScanInPath:
			// input is filled with path data
			scanData = param.Input
		case ScanInQuery:
			// input may not be filled and needs to be parsed from the request
			// TODO: if more than one parameter, and array, call this function with recursion
			scanData = strings.Join(r.URL.Query()[param.Input], " ")
		default:
			panic(fmt.Errorf("Impossible scanning location: %d", param.Location))
		}

		n, err := fmt.Sscan(scanData, param.Data)
		if n != 1 {
			WriteError(w, http.StatusBadRequest, &Error{
				Title: err.Error(),
				Source: &map[string]interface{}{
					"parameter": param.Input,
				},
			})
			return false
		}
	}
	return true
}
