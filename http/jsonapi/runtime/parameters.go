// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/24 by Vincent Landgraf

package runtime

import (
	"fmt"
	"net/http"
	"reflect"
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
	Input string
	// Name of the query variable
	Name string
}

// ScanParameters scans the request using the given path parameter objects
// in case an error is encountered a 400 along with a jsonapi errors object
// is sent to the ResponseWriter and false is returned. Returns true if all
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
			input := r.URL.Query()[param.Name]

			// if parameter is a slice
			reValue := reflect.ValueOf(param.Data).Elem()
			if reValue.Kind() == reflect.Slice {
				size := len(input)
				array := reflect.MakeSlice(reValue.Type(), size, size)
				invalid := 0
				for i := 0; i < size; i++ {
					if input[i] == "" {
						invalid++
						continue
					}

					arrElem := array.Index(i - invalid)
					n, _ := fmt.Sscan(input[i], arrElem.Addr().Interface())
					if n != 1 {
						WriteError(w, http.StatusBadRequest, &Error{
							Title:  fmt.Sprintf("invalid value for %s", param.Name),
							Detail: fmt.Sprintf("invalid value, expected %s got: %q", arrElem.Type(), input[i]),
							Source: &map[string]interface{}{
								"parameter": param.Name,
							},
						})
						return false
					}
				}
				// some of the query parameters where empty, filter them out
				if invalid > 0 {
					array = array.Slice(0, size-invalid)
				}
				reValue.Set(array)

				// skip parsing at the bottom of the loop
				continue
			}

			// single parameter scanning
			scanData = strings.Join(input, " ")
		default:
			panic(fmt.Errorf("Impossible scanning location: %d", param.Location))
		}

		n, _ := fmt.Sscan(scanData, param.Data)
		// only report on non empty data, govalidator will handle the other cases
		if n != 1 && scanData != "" {
			WriteError(w, http.StatusBadRequest, &Error{
				Title:  fmt.Sprintf("invalid value for %s", param.Name),
				Detail: fmt.Sprintf("invalid value, expected %s got: %q", reflect.ValueOf(param.Data).Type(), scanData),
				Source: &map[string]interface{}{
					"parameter": param.Name,
				},
			})
			return false
		}
	}
	return true
}
