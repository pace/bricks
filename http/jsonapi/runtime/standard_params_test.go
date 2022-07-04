// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/02/06 by Charlotte Pröller

package runtime_test

import (
	"context"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/pace/bricks/backend/postgres"
	"github.com/pace/bricks/http/jsonapi/runtime"
	"github.com/pace/bricks/maintenance/log"
	"github.com/stretchr/testify/assert"
)

type TestModel struct {
	FilterName string
}

type testValueSanitizer struct {
}

func (t *testValueSanitizer) SanitizeValue(fieldName string, value string) (interface{}, error) {
	return value, nil
}

func TestIntegrationFilterParameter(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	// Setup
	a := assert.New(t)
	db := setupDatabase(a)
	mappingNames := map[string]string{
		"test": "filter_name",
	}
	mapper := runtime.NewMapMapper(mappingNames)
	// filter
	r := httptest.NewRequest("GET", "http://abc.de/whatEver?filter[test]=b", nil)
	urlParams, err := runtime.ReadURLQueryParameters(r, mapper, &testValueSanitizer{})
	a.NoError(err)
	var modelsFilter []TestModel
	q := db.Model(&modelsFilter)
	q = urlParams.AddToQuery(q)
	count, _ := q.SelectAndCount()
	a.Equal(1, count)
	a.Equal("b", modelsFilter[0].FilterName)

	r = httptest.NewRequest("GET", "http://abc.de/whatEver?filter[test]=a,b", nil)
	urlParams, err = runtime.ReadURLQueryParameters(r, mapper, &testValueSanitizer{})
	a.NoError(err)
	var modelsFilter2 []TestModel
	q = db.Model(&modelsFilter2)
	q = urlParams.AddToQuery(q)
	count, _ = q.SelectAndCount()
	a.Equal(2, count)
	sort.Slice(modelsFilter2, func(i, j int) bool {
		return modelsFilter2[i].FilterName < modelsFilter2[j].FilterName
	})
	a.Equal("a", modelsFilter2[0].FilterName)
	a.Equal("b", modelsFilter2[1].FilterName)

	// Paging
	r = httptest.NewRequest("GET", "http://abc.de/whatEver?page[number]=3&page[size]=2", nil)
	urlParams, err = runtime.ReadURLQueryParameters(r, mapper, &testValueSanitizer{})
	assert.NoError(t, err)
	var modelsPaging []TestModel
	q = db.Model(&modelsPaging)
	q = urlParams.AddToQuery(q)
	err = q.Select()
	a.NoError(err)
	a.Equal(0, len(modelsPaging))

	// Sorting
	r = httptest.NewRequest("GET", "http://abc.de/whatEver?sort=-test", nil)
	urlParams, err = runtime.ReadURLQueryParameters(r, mapper, &testValueSanitizer{})
	assert.NoError(t, err)
	var modelsSort []TestModel
	q = db.Model(&modelsSort)
	q = urlParams.AddToQuery(q)
	err = q.Select()
	a.NoError(err)
	a.Equal(3, len(modelsSort))
	a.Equal("c", modelsSort[0].FilterName)
	a.Equal("b", modelsSort[1].FilterName)
	a.Equal("a", modelsSort[2].FilterName)

	// Combine all
	r = httptest.NewRequest("GET", "http://abc.de/whatEver?sort=-test&filter[test]=a,b&page[number]=0&page[size]=1", nil)
	urlParams, err = runtime.ReadURLQueryParameters(r, mapper, &testValueSanitizer{})
	assert.NoError(t, err)
	var modelsCombined []TestModel
	q = db.Model(&modelsCombined)
	q = urlParams.AddToQuery(q)
	err = q.Select()
	assert.NoError(t, err)
	a.Equal(1, len(modelsCombined))
	a.Equal("b", modelsCombined[0].FilterName)
	// Tear Down
	err = db.DropTable(&TestModel{}, &orm.DropTableOptions{
		IfExists: true,
	})
	assert.NoError(t, err)
}

func setupDatabase(a *assert.Assertions) *pg.DB {
	dB := postgres.DefaultConnectionPool()
	dB = dB.WithContext(log.WithContext(context.Background()))

	a.NoError(dB.CreateTable(&TestModel{}, &orm.CreateTableOptions{IfNotExists: true}))
	_, err := dB.Model(&TestModel{
		FilterName: "a",
	}).Insert()
	a.NoError(err)
	_, err = dB.Model(&TestModel{
		FilterName: "b",
	}).Insert()
	a.NoError(err)
	_, err = dB.Model(&TestModel{
		FilterName: "c",
	}).Insert()
	a.NoError(err)
	return dB
}
