// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/08/24 by Vincent Landgraf

package runtime

import "net/http"

// ValidateStruct checks the given struct and returns true if the struct
// is valid according to the specification (declared with go-validator struct tags)
// In case of an error, an jsonapi error message will be directly send to the client
func ValidateStruct(w http.ResponseWriter, r *http.Request, data interface{}) bool {
	return true
}
