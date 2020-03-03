// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/02/06 by Charlotte Pröller

package runtime

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/caarlos0/env"
	"github.com/go-pg/pg"
	"github.com/pace/bricks/maintenance/log"
)

// QueryOption is a function that applies an option (like sorting, filter or pagination) to a database query
type QueryOption func(query Query) Query

// Query based on orm.Query but only the needed functions
// guaranties testability of all of the following functions
type Query interface {
	Offset(n int) Query
	Limit(n int) Query
	Order(orders ...string) Query
	Where(condition string, params ...interface{}) Query
}

type config struct {
	MaxPageSize int `env:"MAX_PAGE_SIZE" envDefault:"100"`
	MinPageSize int `env:"MIN_PAGE_SIZE" envDefault:"1"`
}

var cfg config

func init() {
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Failed to parse jsonapi params from environment: %v", err)
	}
}

// ValueSanitizer should sanitize query parameter values,
// the implementation should validate the value and transform it to the right type.
type ValueSanitizer interface {
	// SanitizeValue should sanitize a value, that should be in the column fieldName
	SanitizeValue(fieldName string, value string) (interface{}, error)
}

// PagingFromRequest extracts pagination query parameter and  returns a function that adds the pagination to a query
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
	if (pageSize < cfg.MinPageSize) || (pageSize > cfg.MaxPageSize) {
		return nil, fmt.Errorf("invalid pagesize not between min. and max. value, min: %d, max: %d", cfg.MinPageSize, cfg.MaxPageSize)
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
// Database model and response type may differ, so the mapper allows to map the name of  field from the request
// to a database column name
// returns a (possible nop) queryOption, even if any sorting parameter was invalid
// The error contains a list of all invalid parameters. Invalid parameters are not added to the query.
func SortingFromRequest(r *http.Request, modelMapping map[string]string) (QueryOption, error) {
	sort := r.URL.Query().Get("sort")
	if sort == "" {
		return func(query Query) Query { return query }, nil
	}
	sorting := strings.Split(sort, ",")

	var order string
	var resultedOrders []string
	var errSortingWithReason []string
	for _, val := range sorting {
		if val == "" {
			continue
		}
		order = " ASC"
		if strings.HasPrefix(val, "-") {
			order = " DESC"
		}
		val = strings.TrimPrefix(val, "-")

		key, isValid := modelMapping[val]
		if !isValid {
			errSortingWithReason = append(errSortingWithReason, val)
			continue
		}
		resultedOrders = append(resultedOrders, key+order)
	}
	sortingFilterOption := func(query Query) Query {
		for _, val := range resultedOrders {
			query.Order(val)
		}
		return query
	}

	if len(errSortingWithReason) != 0 {
		return sortingFilterOption, fmt.Errorf("at least one sorting parameter is not valid: %q", strings.Join(errSortingWithReason, ","))
	}
	return sortingFilterOption, nil
}

// FilterFromRequest adds filter to a query based on the request query parameter
// filter[name]=val1,val2 results in name IN (val1, val2), filter[name]=val results in name=val
// Database model and response type may differ, so the mapper allows to map the name of field from the request
// to a database column name and the sanitizer allows to correct type of the value and sanitize it.
// Will always return a QueryOptions function with all valid filters (can be a nop)
// if any filter are invalid a error with a list of all invalid filters is returned
func FilterFromRequest(r *http.Request, modelMapping map[string]string, sanitizer ValueSanitizer) (QueryOption, error) {
	filter := make(map[string][]interface{})
	var invalidFilter []string
	for queryName, queryValues := range r.URL.Query() {
		if !(strings.HasPrefix(queryName, "filter[") && strings.HasSuffix(queryName, "]")) {
			continue
		}
		key, isValid := getFilterKey(queryName, modelMapping)
		if !isValid {
			invalidFilter = append(invalidFilter, key)
			continue
		}
		filterValues, isValid := getFilterValues(key, queryValues, sanitizer)
		if !isValid {
			invalidFilter = append(invalidFilter, key)
			continue
		}
		filter[key] = filterValues
	}

	filterQueryOption := func(query Query) Query {
		for name, filterValues := range filter {
			if len(filterValues) == 0 {
				continue
			}

			if len(filterValues) == 1 {
				query.Where("?=?", name, filterValues[0])
				continue
			}
			query.Where("? IN (?)", name, pg.In(filterValues))
		}
		return query
	}

	if len(invalidFilter) != 0 {
		return filterQueryOption, fmt.Errorf("at least one filter parameter is not valid: %q", strings.Join(invalidFilter, ","))
	}
	return filterQueryOption, nil
}

func getFilterKey(queryName string, modelMapping map[string]string) (string, bool) {
	field := strings.TrimPrefix(queryName, "filter[")
	field = strings.TrimSuffix(field, "]")
	mapped, isValid := modelMapping[field]
	if !isValid {
		return field, false
	}
	return mapped, true
}

func getFilterValues(fieldName string, queryValues []string, sanitizer ValueSanitizer) ([]interface{}, bool) {
	var filterValues []interface{}
	for _, value := range queryValues {
		separatedValues := strings.Split(value, ",")
		for _, separatedValue := range separatedValues {
			sanitized, err := sanitizer.SanitizeValue(fieldName, separatedValue)
			if err != nil {
				return nil, false
			}
			filterValues = append(filterValues, sanitized)
		}
	}
	return filterValues, true
}
