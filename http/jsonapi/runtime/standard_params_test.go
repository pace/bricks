// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package runtime_test

import (
	"context"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/stretchr/testify/assert"

	"github.com/pace/bricks/backend/postgres"
	"github.com/pace/bricks/http/jsonapi/runtime"
	"github.com/pace/bricks/maintenance/log"
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
	defer func() {
		// Tear Down
		err := db.DropTable(&TestModel{}, &orm.DropTableOptions{})
		assert.NoError(t, err)
	}()

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
	r = httptest.NewRequest("GET", "http://abc.de/whatEver?page[number]=1&page[size]=2", nil)
	urlParams, err = runtime.ReadURLQueryParameters(r, mapper, &testValueSanitizer{})
	assert.NoError(t, err)
	var modelsPaging []TestModel
	q = db.Model(&modelsPaging)
	q = urlParams.AddToQuery(q)
	err = q.Select()
	a.NoError(err)
	sort.Slice(modelsPaging, func(i, j int) bool {
		return modelsPaging[i].FilterName < modelsPaging[j].FilterName
	})
	a.Equal("c", modelsPaging[0].FilterName)
	a.Equal("d", modelsPaging[1].FilterName)

	// Sorting
	r = httptest.NewRequest("GET", "http://abc.de/whatEver?sort=-test", nil)
	urlParams, err = runtime.ReadURLQueryParameters(r, mapper, &testValueSanitizer{})
	assert.NoError(t, err)
	var modelsSort []TestModel
	q = db.Model(&modelsSort)
	q = urlParams.AddToQuery(q)
	err = q.Select()
	a.NoError(err)
	a.Equal(6, len(modelsSort))
	a.Equal("f", modelsSort[0].FilterName)
	a.Equal("e", modelsSort[1].FilterName)
	a.Equal("d", modelsSort[2].FilterName)
	a.Equal("c", modelsSort[3].FilterName)
	a.Equal("b", modelsSort[4].FilterName)
	a.Equal("a", modelsSort[5].FilterName)

	// Combine all
	r = httptest.NewRequest("GET", "http://abc.de/whatEver?sort=-test&filter[test]=a,b,e,f&page[number]=1&page[size]=2", nil)
	urlParams, err = runtime.ReadURLQueryParameters(r, mapper, &testValueSanitizer{})
	assert.NoError(t, err)
	var modelsCombined []TestModel
	q = db.Model(&modelsCombined)
	q = urlParams.AddToQuery(q)
	err = q.Select()
	assert.NoError(t, err)
	a.Equal(2, len(modelsCombined))
	a.Equal("b", modelsCombined[0].FilterName)
	a.Equal("a", modelsCombined[1].FilterName)
}

func setupDatabase(a *assert.Assertions) *pg.DB {
	db := postgres.DefaultConnectionPool()
	db = db.WithContext(log.WithContext(context.Background()))

	err := db.CreateTable(&TestModel{}, &orm.CreateTableOptions{})
	a.NoError(err)
	_, err = db.Model(&TestModel{
		FilterName: "a",
	}).Insert()
	a.NoError(err)

	_, err = db.Model(&TestModel{
		FilterName: "b",
	}).Insert()
	a.NoError(err)

	_, err = db.Model(&TestModel{
		FilterName: "c",
	}).Insert()
	a.NoError(err)

	_, err = db.Model(&TestModel{
		FilterName: "d",
	}).Insert()
	a.NoError(err)

	_, err = db.Model(&TestModel{
		FilterName: "e",
	}).Insert()
	a.NoError(err)

	_, err = db.Model(&TestModel{
		FilterName: "f",
	}).Insert()
	a.NoError(err)

	return db
}
