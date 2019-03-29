// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/03/29 by Florian Hübsch

package oauth2

import (
	"fmt"
	"index/suffixarray"
	"regexp"
	"strings"
)

// Scope represents an OAuth 2 access token scope
type Scope string

// IsIncludedIn checks if the permissions of a scope s are also included
// in the provided scope t. This can be useful to check if a scope has all
// required permissions to access an endpoint.
func (s *Scope) IsIncludedIn(t Scope) bool {
	// build index for substring search in logarithmic time
	index := suffixarray.New([]byte(t))

	// permission list of scope s
	ps := s.toSlice()

	for _, p := range ps {
		expr := fmt.Sprintf(" %s | %s$|^%s ", p, p, p)
		r, _ := regexp.Compile(expr)
		res := index.FindAllIndex(r, -1)

		// return false if permission not found in scope t
		if len(res) == 0 {
			return false
		}
	}

	return true
}

func (s *Scope) toSlice() []string {
	return strings.Fields(string(*s))
}
