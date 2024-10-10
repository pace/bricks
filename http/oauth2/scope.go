// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.

package oauth2

import (
	"slices"
	"strings"
)

// Scope represents an OAuth 2 access token scope.
type Scope string

// IsIncludedIn checks if the permissions of a scope s are also included
// in the provided scope t. This can be useful to check if a scope has all
// required permissions to access an endpoint.
func (s *Scope) IsIncludedIn(t Scope) bool {
	// permission list of scope s
	pss := s.toSlice()
	// permission list of scope t
	pts := t.toSlice()

	for _, ps := range pss {
		if !slices.Contains(pts, ps) {
			return false
		}
	}

	return true
}

func (s *Scope) toSlice() []string {
	return strings.Fields(string(*s))
}

func (s *Scope) Add(scope string) Scope {
	return Scope(strings.Join(append(s.toSlice(), scope), " "))
}
