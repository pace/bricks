// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/24 by Vincent Landgraf

package runtime

import (
	"net/http"
)

// Unmarshal processes the request content and fills passed data struct with the
// correct jsonapi content. After un-marshaling the struct will be validated with
// specified go-validator struct tags.
// In case of an error, an jsonapi error message will be directly send to the client
func Unmarshal(w http.ResponseWriter, r *http.Request, data interface{}) bool {
	// go validator
	// 422 if the request body is not okay

	return true
}

// Marshal the given data and writes them into the response writer, sets
// the content-type and code as well
func Marshal(w http.ResponseWriter, data interface{}, code int) {
}
