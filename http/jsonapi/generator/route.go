// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package generator

import (
	"net/url"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

type route struct {
	method, pattern, handler, serviceFunc       string
	requestType, responseType, responseTypeImpl string
	operation                                   *openapi3.Operation
	url                                         *url.URL
	queryValues                                 url.Values
}

type sortableRouteList []*route

func (r *route) parseURL() (err error) {
	r.url, err = url.Parse(r.pattern)
	if err != nil {
		return
	}
	r.queryValues = r.url.Query() // cache query values
	return
}

// Len is the number of elements in the collection.
func (l *sortableRouteList) Len() int {
	return len(*l)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (l *sortableRouteList) Less(i, j int) bool {
	elemI, elemJ := (*l)[i], (*l)[j]

	// Prio, generic to the bottom, specific to the top (less):
	// * longer paths are more specific (paths separated by "/" or ".")
	// * no parameter values is more specific
	// * more query values are more specific
	if a, b := pathLen(elemI.url.Path), pathLen(elemJ.url.Path); a != b {
		return a > b
	}
	if a, b := strings.Count(elemJ.url.Path, "{"), strings.Count(elemI.url.Path, "{"); a != b {
		return a > b
	}
	return len(elemI.queryValues) > len(elemJ.queryValues)
}

// Swap swaps the elements with indexes i and j.
func (l *sortableRouteList) Swap(i, j int) {
	(*l)[i], (*l)[j] = (*l)[j], (*l)[i]
}

func pathLen(in string) int {
	acc := 0
	for _, sep := range []string{"/", "."} {
		acc += strings.Count(in, sep)
	}

	return acc
}
