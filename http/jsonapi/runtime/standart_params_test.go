// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/02/06 by Charlotte Pröller

package runtime

/*
type testQuery struct {
	offset int
	limit  int
	orders []string
	where  []string
}

func (q *testQuery) Where(condition string, params ...interface{}) *testQuery {
	where := condition
	for _, param := range params {
		if paramString, ok := param.(string); ok {
			where = strings.Replace(where, "?", paramString, 1)
		}
		if paramTypeVal, ok := param.(types.ValueAppender); ok {
			var value []byte
			value = paramTypeVal.AppendValue(value, 0)
			where = strings.Replace(where, "?", string(value), 1)
		}

	}
	q.where = append(q.where, where)
	return q
}

func (q *testQuery) Offset(n int) *testQuery {
	q.offset = n
	return q
}

func (q *testQuery) Limit(n int) *testQuery {
	q.limit = n
	return q
}

func (q *testQuery) Order(orders ...string) *testQuery {
	q.orders = append(q.orders, orders...)
	return q
}

type testSanitizer struct {
}

type testModel struct {

}



func (t testDb) FormatQuery(b []byte, query string, params ...interface{}) []byte {
	panic("implement me")
}

func (t testSanitizer) SanitizeValue(fieldName string, value string) (interface{}, error) {
	return value + "san", nil
}

func TestPagingFromRequest(t *testing.T) {
	r := httptest.NewRequest("GET", "/articles?page[number]=3&page[size]=1", nil)
	queryOpt, err := PagingFromRequest(r)
	require.NotNil(t, queryOpt)
	require.NoError(t,err)
	q := orm.NewQuery(testdb, &testModel{})
	q.Apply(queryOpt)

	orm.Formatter
	require.Equal(t, 2, result.offset)
	require.Equal(t, 1, result.limit)
	require.Equal(t, 0, len(result.orders))
}

func TestPagingMaxMin(t *testing.T) {
	r := httptest.NewRequest("GET", "/articles?page[number]=3&page[size]=0", nil)
	_, err := PagingFromRequest(r)
	require.Error(t, err)

	r = httptest.NewRequest("GET", "/articles?page[number]=3&page[size]=200", nil)
	_, err = PagingFromRequest(r)
	require.Error(t, err)

}

func TestSortingFromRequest(t *testing.T) {
	r := httptest.NewRequest("GET", "/articles?sort=abc,-def,qwe", nil)
	mapper := map[string]string{
		"abc": "abc",
		"def": "def",
		"qwe": "qwe",
	}
	queryOpt, err := SortingFromRequest(r, mapper)
	res := queryOpt(&testQuery{})
	require.NotNil(t, queryOpt)
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

func TestMapperSorting(t *testing.T) {
	r := httptest.NewRequest("GET", "/articles?sort=abc,-def", nil)
	mapper := map[string]string{
		"def": "def",
	}
	queryOpt, err := SortingFromRequest(r, mapper)
	require.NotNil(t, queryOpt)
	res := queryOpt(&testQuery{})
	result := res.(*testQuery)
	require.Error(t, err)
	require.Equal(t, `at least one sorting parameter is not valid: "abc"`, err.Error())
	require.Equal(t, 0, result.offset)
	require.Equal(t, 0, result.limit)
	require.Equal(t, 1, len(result.orders))
	require.Equal(t, 0, len(result.where))
	require.Contains(t, result.orders[0], "def DESC")
}

func TestMapperFiltering(t *testing.T) {
	r := httptest.NewRequest("GET", "/articles?filter[abc]=1,2&filter[name]=1", nil)
	mapper := map[string]string{
		"abc": "abc",
	}

	queryOpt, err := FilterFromRequest(r, mapper, &testSanitizer{})
	res := queryOpt(&testQuery{})
	require.NotNil(t, queryOpt)
	result := res.(*testQuery)
	require.Error(t, err)
	require.Equal(t, `at least one filter parameter is not valid: "name"`, err.Error())
	require.Equal(t, 0, result.offset)
	require.Equal(t, 0, result.limit)
	require.Equal(t, 0, len(result.orders))
	require.Equal(t, 1, len(result.where))
	require.Contains(t, result.where[0], "abc IN (1san,2san)")
}

func TestFilteringFromRequest(t *testing.T) {
	r := httptest.NewRequest("GET", "/articles?filter[abc]=1,2&filter[name]=1&testy=test", nil)
	mapper := map[string]string{
		"abc":   "abc",
		"name":  "name",
		"testy": "testy",
	}
	queryOpt, err := FilterFromRequest(r, mapper, &testSanitizer{})
	res := queryOpt(&testQuery{})
	require.NotNil(t, queryOpt)
	result := res.(*testQuery)
	require.NoError(t, err)
	require.Equal(t, 0, result.offset)
	require.Equal(t, 0, result.limit)
	require.Equal(t, 0, len(result.orders))
	require.Equal(t, 2, len(result.where))
	for _, val := range result.where {
		require.False(t, strings.Contains(val, "?"))
	}
	require.Contains(t, result.where, "abc IN (1san,2san)")
	require.Contains(t, result.where, "name=1san")
}

func TestAllInOne(t *testing.T) {
	r := httptest.NewRequest("GET", "/articles?filter[abc]=1,2&filter[name]=1&testy=test&sort=abc,-def&page[number]=3&page[size]=1", nil)
	mapper := map[string]string{
		"abc":   "abc",
		"name":  "name",
		"testy": "testy",
		"def":   "def",
	}
	queryOpt, err := FilterFromRequest(r, mapper, &testSanitizer{})
	require.NoError(t, err)
	res := queryOpt(&testQuery{})
	queryOpt, err = SortingFromRequest(r, mapper)
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
	require.Contains(t, result.where, "abc IN (1san,2san)")
	require.Contains(t, result.where, "name=1san")
	require.Contains(t, result.orders, "abc ASC")
	require.Contains(t, result.orders, "def DESC")
}

*/
