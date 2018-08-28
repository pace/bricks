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
			input := r.URL.Query()[param.Input]

			// if parameter is a slice
			reValue := reflect.ValueOf(param.Data).Elem()
			if reValue.Kind() == reflect.Slice {
				size := len(input)
				array := reflect.MakeSlice(reValue.Type(), size, size)
				for i := 0; i < size; i++ {
					n, _ := fmt.Sscan(input[i], array.Index(i).Addr().Interface())
					if n != 1 {
						WriteError(w, http.StatusBadRequest, &Error{
							Title: fmt.Sprintf("invalid value, exepcted %s got: %q", array.Index(i).Type(), input[i]),
							Source: &map[string]interface{}{
								"parameter": param.Input,
							},
						})
						return false
					}
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
