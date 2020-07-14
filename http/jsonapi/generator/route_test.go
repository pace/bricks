// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/07/14 by Marius Neugebauer

package generator

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSortableRouteList(t *testing.T) {
	paths := []string{
		"/beta/payment-methods/{id}",
		"/beta/payment-methods",
		"/beta/payment-methods/longer",
		"/beta/payment-methods?filter[status]=valid",
	}
	list := make(sortableRouteList, len(paths))
	for i, path := range paths {
		route := &route{pattern: path}
		require.NoError(t, route.parseURL())
		list[i] = route
	}
	sort.Sort(&list)
	actual := make([]string, len(paths))
	for i, route := range list {
		actual[i] = route.pattern
	}
	assert.Equal(t, []string{
		"/beta/payment-methods/longer",
		"/beta/payment-methods/{id}",
		"/beta/payment-methods?filter[status]=valid",
		"/beta/payment-methods",
	}, actual)
}
