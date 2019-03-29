// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/03/29 by Florian Hübsch

package oauth2

import (
	"strings"
)

// Scope represents an OAuth 2 access token scope
type Scope string

// IsIncludedIn checks if the permissions of a scope s are also included
// in the provided scope t. This can be useful to check if a scope has all
// required permissions to access an endpoint.
func (s *Scope) IsIncludedIn(t Scope) bool {
	// permission list of scope s
	pss := s.toSlice()
	// permission list of scope t
	pts := t.toSlice()

	var found bool

	for _, ps := range pss {
		found = false

		for _, pt := range pts {
			if ps == pt {
				found = true
				break
			}
		}

		if found == false {
			return false
		}
	}

	return true
}

func (s *Scope) toSlice() []string {
	return strings.Fields(string(*s))
}
