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
	"github.com/go-pg/pg/orm"
	"github.com/pace/bricks/maintenance/log"
)

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

// ColumnMapper maps the name of a filter or sorting parameter to a database column name
type ColumnMapper interface {
	// Map maps the value, this function decides if the value is allowed and translates it to a database column name,
	// the function returns the database column name and a bool that indicates that the value is allowed and mapped
	Map(value string) (string, bool)
}

// MapMapper is a very easy ColumnMapper implementation based on a map which contains all allowed values
// and maps them with a map
type MapMapper struct {
	mapping map[string]string
}

// NewMapMapper returns a MapMapper for a specific map
func NewMapMapper(mapping map[string]string) *MapMapper {
	return &MapMapper{mapping: mapping}
}

// Map returns the mapped value and if it is valid based on a map
func (m *MapMapper) Map(value string) (string, bool) {
	val, isValid := m.mapping[value]
	return val, isValid
}

// UrlQueryParameters contains all information that are needed for pagination, sorting and filtering.
// It is not depending on orm.Query
type UrlQueryParameters struct {
	HasPagination bool
	PageNr        int
	PageSize      int
	Order         []string
	Filter        map[string][]interface{}
}

// ReadURLQueryParameters reads sorting, filter and pagination from requests and return a UrlQueryParameters object,
// even if any errors occur. The returned error combines all errors of pagination, filter and sorting.
func ReadURLQueryParameters(r *http.Request, mapper ColumnMapper, sanitizer ValueSanitizer) (*UrlQueryParameters, error) {
	result := &UrlQueryParameters{}
	errPagination := result.setPagination(r)
	errSorting := result.setSorting(r, mapper)
	errFilter := result.setFilter(r, mapper, sanitizer)
	if errPagination != nil || errSorting != nil || errFilter != nil {
		err := fmt.Errorf("problems occured while ready filter, sorting or pagination from request: filter: %w, sorting: %w, pagination: %w", errFilter, errSorting, errPagination)
		return result, err
	}
	return result, nil
}

// AddToQuery adds filter, sorting and pagination to a orm.Query
func (u *UrlQueryParameters) AddToQuery(query *orm.Query) *orm.Query {
	if u.HasPagination {
		if u.PageNr == 0 {
			query.Offset(0)
		} else {
			query.Offset((u.PageSize * u.PageNr) - 1)
		}
		query.Limit(u.PageSize)
	}
	for name, filterValues := range u.Filter {
		if len(filterValues) == 0 {
			continue
		}

		if len(filterValues) == 1 {
			query.Where(name+" = ?", filterValues[0])
			continue
		}
		query.Where(name+" IN (?)", pg.In(filterValues))
	}
	for _, val := range u.Order {
		query.Order(val)
	}
	return query

}

func (u *UrlQueryParameters) setPagination(r *http.Request) error {
	pageStr := r.URL.Query().Get("page[number]")
	sizeStr := r.URL.Query().Get("page[size]")
	if pageStr == "" || sizeStr == "" {
		u.HasPagination = false
		return nil
	}
	u.HasPagination = true
	pageNr, err := strconv.Atoi(pageStr)
	if err != nil {
		return err
	}
	pageSize, err := strconv.Atoi(sizeStr)
	if err != nil {
		return err
	}
	if (pageSize < cfg.MinPageSize) || (pageSize > cfg.MaxPageSize) {
		return fmt.Errorf("invalid pagesize not between min. and max. value, min: %d, max: %d", cfg.MinPageSize, cfg.MaxPageSize)
	}
	u.PageNr = pageNr
	u.PageSize = pageSize
	return nil
}

func (u *UrlQueryParameters) setSorting(r *http.Request, mapper ColumnMapper) error {
	sort := r.URL.Query().Get("sort")
	if sort == "" {
		return nil
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

		key, isValid := mapper.Map(val)
		if !isValid {
			errSortingWithReason = append(errSortingWithReason, val)
			continue
		}
		resultedOrders = append(resultedOrders, key+order)
	}
	u.Order = resultedOrders
	if len(errSortingWithReason) > 0 {
		return fmt.Errorf("at least one sorting parameter is not valid: %q", strings.Join(errSortingWithReason, ","))
	}
	return nil
}

func (u *UrlQueryParameters) setFilter(r *http.Request, mapper ColumnMapper, sanitizer ValueSanitizer) error {
	filter := make(map[string][]interface{})
	var invalidFilter []string
	for queryName, queryValues := range r.URL.Query() {
		if !(strings.HasPrefix(queryName, "filter[") && strings.HasSuffix(queryName, "]")) {
			continue
		}
		key, isValid := getFilterKey(queryName, mapper)
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
	u.Filter = filter
	if len(invalidFilter) != 0 {
		return fmt.Errorf("at least one filter parameter is not valid: %q", strings.Join(invalidFilter, ","))
	}
	return nil
}

func getFilterKey(queryName string, modelMapping ColumnMapper) (string, bool) {
	field := strings.TrimPrefix(queryName, "filter[")
	field = strings.TrimSuffix(field, "]")
	mapped, isValid := modelMapping.Map(field)
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
