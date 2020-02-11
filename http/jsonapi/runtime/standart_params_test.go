// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/02/06 by Charlotte Pröller

package runtime

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type testQuery struct {
	offset int
	limit  int
	orders []string
	where  []string
}

func (q *testQuery) Where(condition string, params ...interface{}) Query {
	var replacerVals []string
	for _, param := range params {
		replacerVals = append(replacerVals, "?", param.(string))
	}
	q.where = append(q.where, strings.NewReplacer(replacerVals...).Replace(condition))
	return q
}

func (q *testQuery) Offset(n int) Query {
	q.offset = n
	return q
}

func (q *testQuery) Limit(n int) Query {
	q.limit = n
	return q
}

func (q *testQuery) Order(orders ...string) Query {
	q.orders = append(q.orders, orders...)
	return q
}

func TestPagingFromRequest(t *testing.T) {
	r := httptest.NewRequest("GET", "/articles?page[number]=3&page[size]=1", nil)
	queryOpt, err := PagingFromRequest(r)
	res := queryOpt(&testQuery{})
	result := res.(*testQuery)
	require.NoError(t, err)
	require.Equal(t, 2, result.offset)
	require.Equal(t, 1, result.limit)
	require.Equal(t, 0, len(result.orders))
}

func TestSortingFromRequest(t *testing.T) {
	r := httptest.NewRequest("GET", "/articles?sort=abc,-def,qwe", nil)
	queryOpt, err := SortingFromRequest(r, DefaultMapper)
	res := queryOpt(&testQuery{})
	result := res.(*testQuery)
	require.NoError(t, err)
	require.Equal(t, 0, result.offset)
	require.Equal(t, 0, result.limit)
	require.Equal(t, 3, len(result.orders))
	require.Equal(t, 0, len(result.where))
	require.Contains(t, result.orders[0], "abc ASC")
	require.Contains(t, result.orders[1], "def DESC")
	require.Contains(t, result.orders[2], "qwe ASC")
}

func TestFilteringFromRequest(t *testing.T) {
	r := httptest.NewRequest("GET", "/articles?filter[abc]=1,2&filter[name]=1&testy=test", nil)
	queryOpt, err := FilterFromRequest(r, DefaultMapper)
	res := queryOpt(&testQuery{})
	result := res.(*testQuery)
	require.NoError(t, err)
	require.Equal(t, 0, result.offset)
	require.Equal(t, 0, result.limit)
	require.Equal(t, 0, len(result.orders))
	require.Equal(t, 2, len(result.where))
	for _, val := range result.where {
		require.False(t, strings.Contains(val, "?"))
	}
	require.Contains(t, result.where[0], "abc IN (1,2)")
	require.Contains(t, result.where[1], "name=1")
}

func TestAllInOne(t *testing.T) {
	r := httptest.NewRequest("GET", "/articles?filter[abc]=1,2&filter[name]=1&testy=test&sort=abc,-def&page[number]=3&page[size]=1", nil)
	queryOpt, err := FilterFromRequest(r, DefaultMapper)
	require.NoError(t, err)
	res := queryOpt(&testQuery{})
	queryOpt, err = SortingFromRequest(r, DefaultMapper)
	require.NoError(t, err)
	res = queryOpt(res)
	queryOpt, err = PagingFromRequest(r)
	require.NoError(t, err)
	res = queryOpt(res)

	result := res.(*testQuery)
	require.NoError(t, err)
	require.Equal(t, 2, result.offset)
	require.Equal(t, 1, result.limit)
	require.Equal(t, 2, len(result.orders))
	require.Equal(t, 2, len(result.where))
	for _, val := range result.where {
		require.False(t, strings.Contains(val, "?"))
	}
	require.Contains(t, result.where, "abc IN (1,2)")
	require.Contains(t, result.where, "name=1")
	require.Contains(t, result.orders, "abc ASC")
	require.Contains(t, result.orders, "def DESC")
}
