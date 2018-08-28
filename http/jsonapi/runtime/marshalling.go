// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/24 by Vincent Landgraf

package runtime

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/google/jsonapi"
)

// Unmarshal processes the request content and fills passed data struct with the
// correct jsonapi content. After un-marshaling the struct will be validated with
// specified go-validator struct tags.
// In case of an error, an jsonapi error message will be directly send to the client
func Unmarshal(w http.ResponseWriter, r *http.Request, data interface{}) bool {
	defer r.Body.Close() // don't leak

	// verify that the client accepts our response
	// Note: logically this would be done before marshalling,
	//       to prevent stale backend/frontend state we respond before
	// 		 Additionally, marshal has no access to the request struct
	accept := r.Header.Get("Accept")
	if accept != jsonAPIContentType {
		WriteError(w, http.StatusNotAcceptable,
			fmt.Errorf("Request needs to be send with %q header, containing value: %q", "Accept", jsonAPIContentType))
		return false
	}

	// if the client didn't send a content type, don't verify
	contentType := r.Header.Get("Content-Type")
	if contentType != jsonAPIContentType {
		WriteError(w, http.StatusUnsupportedMediaType,
			fmt.Errorf("Request needs to be send with %q header, containing value: %q", "Accept", jsonAPIContentType))
		return false
	}

	// parse request
	err := jsonapi.UnmarshalPayload(r.Body, data)
	if err != nil {
		WriteError(w, http.StatusUnprocessableEntity,
			fmt.Errorf("Can't parse content: %v", err))
		return false
	}

	// validate request
	return ValidateRequest(w, r, data)
}

// Marshal the given data and writes them into the response writer, sets
// the content-type and code as well
func Marshal(w http.ResponseWriter, data interface{}, code int) {
	// write response header
	w.Header().Set("Content-Type", jsonAPIContentType)
	w.WriteHeader(code)

	// write marshaled response body
	err := jsonapi.MarshalPayload(w, data)
	if err != nil {
		panic(fmt.Errorf("Failed to marshal jsonapi response for %#v: %s", reflect.ValueOf(data), err))
	}
}
