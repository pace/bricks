// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/02/06 by Charlotte Pröller

package runtime

import (
	"net/http"
	"strconv"
	"strings"
)

// QueryOption is a function that applies an option (like sorting, filter or pagination) to a database query
type QueryOption func(query Query) Query

// Query based on orm.Query but only the needed function
// guaranties testability of all of the following functions
type Query interface {
	Offset(n int) Query
	Limit(n int) Query
	Order(orders ...string) Query
	Where(condition string, params ...interface{}) Query
}

type Mapper func(string) string

func PagingFromRequest(r *http.Request) (QueryOption, error) {
	pageStr := r.URL.Query().Get("page[number]")
	sizeStr := r.URL.Query().Get("page[size]")
	if pageStr == "" || sizeStr == "" {
		return func(query Query) Query { return query }, nil
	}

	pageNr, err := strconv.Atoi(pageStr)
	if err != nil {
		return nil, err
	}
	pageSize, err := strconv.Atoi(sizeStr)
	if err != nil {
		return nil, err
	}
	return func(query Query) Query {
		if pageNr == 0 {
			query.Offset(0)
		} else {
			query.Offset(pageSize * (pageNr - 1))
		}
		query.Limit(pageSize)
		return query
	}, nil
}

// SortingFromRequest adds sorting to query based on the request query parameter
// Database model and response type may differ, so the mapper function allows to map the name of  field from the request
// to a database column name
func SortingFromRequest(r *http.Request, mapper Mapper) (QueryOption, error) {
	sort := r.URL.Query().Get("sort")
	if sort == "" {
		return func(query Query) Query { return query }, nil
	}
	sorting := strings.Split(sort, ",")

	return func(query Query) Query {
		var order string
		for _, val := range sorting {
			if val == "" {
				continue
			}
			order = " ASC"
			if strings.HasPrefix(val, "-") {
				order = " DESC"
			}
			val = strings.TrimPrefix(val, "-")
			query.Order(mapper(val) + order)
		}
		return query
	}, nil
}

// FilterFromRequest adds filter to a query based on the request query parameter
// filter[name]=val1,val2 results in name IN (val1, val2), filter[name]=val results in name=val
// Database model and response type may differ, so the mapper function allows to map the name of  field from the request
// to a database column name
func FilterFromRequest(r *http.Request, mapper Mapper) (QueryOption, error) {
	filter := make(map[string][]string)
	for queryName, val := range r.URL.Query() {
		if !strings.HasPrefix(queryName, "filter") {
			continue
		}
		field := strings.TrimPrefix(queryName, "filter[")
		field = strings.TrimSuffix(field, "]")
		filter[mapper(field)] = val

	}

	return func(query Query) Query {
		for name, vals := range filter {
			if len(vals) == 0 {
				continue
			}
			var filterValues []string
			for _, val := range vals {
				filterValues = append(filterValues, strings.Split(val, ",")...)
			}
			if len(filterValues) == 1 {
				query.Where(name+"=?", filterValues[0])
				continue
			}
			query.Where(name+" IN (?)", strings.Join(filterValues, ","))
		}
		return query
	}, nil
}

// DefaultMapper maps an query field directly name to the same name as database column name
func DefaultMapper(in string) string {
	return in
}
