// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/24 by Vincent Landgraf

package runtime

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/pace/bricks/maintenance/log"
)

// Note: we don't use the jsonapi.ErrorObject because it doesn't implement the
// error interface

// Error objects provide additional information about problems
// encountered while performing an operation.
type Error struct {
	// ID is a unique identifier for this particular occurrence of a problem.
	ID string `json:"id,omitempty"`

	// Title is a short, human-readable summary of the problem that SHOULD NOT change from occurrence to occurrence of the problem, except for purposes of localization.
	Title string `json:"title,omitempty"`

	// Detail is a human-readable explanation specific to this occurrence of the problem. Like title, this field’s value can be localized.
	Detail string `json:"detail,omitempty"`

	// Status is the HTTP status code applicable to this problem, expressed as a string value.
	Status string `json:"status,omitempty"`

	// Code is an application-specific error code, expressed as a string value.
	Code string `json:"code,omitempty"`

	// Source an object containing references to the source of the error, optionally including any of the following members:
	Source *map[string]interface{} `json:"source,omitempty"`

	// Meta is an object containing non-standard meta-information about the error.
	Meta *map[string]interface{} `json:"meta,omitempty"`
}

// setHttpStatus sets the http status for the error object
func (e *Error) setHTTPStatus(code int) {
	e.Status = strconv.Itoa(code)
}

// Error implements the error interface
func (e Error) Error() string {
	return e.Title
}

// Errors is a list of errors
type Errors []*Error

// Error implements the error interface
func (e Errors) Error() string {
	messages := make([]string, len(e))
	for i, err := range e {
		messages[i] = err.Error()
	}
	return strings.Join(messages, "\n")
}

// setHttpStatus sets the http status for the error object
func (e Errors) setHTTPStatus(code int) {
	status := strconv.Itoa(code)
	for _, err := range e {
		err.Status = status
	}
}

// setID sets the error id on the request
func (e Errors) setID(errorID string) {
	for _, err := range e {
		err.ID = errorID
	}
}

// WriteError writes a jsonapi error message to the client
func WriteError(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-Type", JSONAPIContentType)
	w.WriteHeader(code)

	// convert error type for marshaling
	var errList errorObjects

	switch v := err.(type) {
	case Error:
		errList.List = append(errList.List, &v)
	case *Error:
		errList.List = append(errList.List, v)
	case Errors:
		errList.List = v
	default:
		errList.List = []*Error{
			&Error{Title: err.Error()},
		}
	}

	reqID := w.Header().Get("Request-Id")

	// update the http status code of the error
	errList.List.setHTTPStatus(code)

	// log all errors send to clients
	log.Logger().Debug().Str("req_id", reqID).
		Err(errList.List).Msg("error sent to client")
	errList.List.setID(reqID)

	// render the error to the client
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	err = enc.Encode(errList)
	if err != nil {
		log.Logger().Info().Str("req_id", reqID).
			Err(err).Msg("Unable to send error response to the client")
	}
}

// Error objects MUST be returned as an array keyed by errors in the top level of a JSON API document.
type errorObjects struct {
	List Errors `json:"errors"`
}
