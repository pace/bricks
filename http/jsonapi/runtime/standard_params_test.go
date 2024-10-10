// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package runtime_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"

	"github.com/pace/bricks/backend/postgres"
	"github.com/pace/bricks/http/jsonapi/runtime"
)

type TestModel struct {
	FilterName string
}

type testValueSanitizer struct{}

func (t *testValueSanitizer) SanitizeValue(fieldName string, value string) (any, error) {
	return value, nil
}

func TestIntegrationFilterParameter(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	ctx := context.Background()

	// Setup
	db := setupDatabase(ctx, t)

	defer func() {
		// Tear Down
		_, err := db.NewDropTable().Model((*TestModel)(nil)).IfExists().Exec(context.Background())
		require.NoError(t, err)
	}()

	mappingNames := map[string]string{
		"test": "filter_name",
	}
	mapper := runtime.NewMapMapper(mappingNames)

	// filter
	r := httptest.NewRequest(http.MethodGet, "http://abc.de/whatEver?filter[test]=b", nil)

	urlParams, err := runtime.ReadURLQueryParameters(r, mapper, &testValueSanitizer{})
	require.NoError(t, err)

	var modelsFilter []TestModel

	q := db.NewSelect().Model(&modelsFilter)
	q = urlParams.AddToQuery(q)

	count, err := q.ScanAndCount(ctx)
	require.NoError(t, err)

	assert.Equal(t, 1, count)
	assert.Equal(t, "b", modelsFilter[0].FilterName)

	r = httptest.NewRequest(http.MethodGet, "http://abc.de/whatEver?filter[test]=a,b", nil)

	urlParams, err = runtime.ReadURLQueryParameters(r, mapper, &testValueSanitizer{})
	require.NoError(t, err)

	var modelsFilter2 []TestModel

	q = db.NewSelect().Model(&modelsFilter2)
	q = urlParams.AddToQuery(q)

	count, err = q.ScanAndCount(ctx)
	require.NoError(t, err)

	assert.Equal(t, 2, count)

	sort.Slice(modelsFilter2, func(i, j int) bool {
		return modelsFilter2[i].FilterName < modelsFilter2[j].FilterName
	})

	assert.Equal(t, "a", modelsFilter2[0].FilterName)
	assert.Equal(t, "b", modelsFilter2[1].FilterName)

	// Paging
	r = httptest.NewRequest(http.MethodGet, "http://abc.de/whatEver?page[number]=1&page[size]=2", nil)

	urlParams, err = runtime.ReadURLQueryParameters(r, mapper, &testValueSanitizer{})
	require.NoError(t, err)

	var modelsPaging []TestModel

	q = db.NewSelect().Model(&modelsPaging)
	q = urlParams.AddToQuery(q)

	err = q.Scan(ctx)
	require.NoError(t, err)

	sort.Slice(modelsPaging, func(i, j int) bool {
		return modelsPaging[i].FilterName < modelsPaging[j].FilterName
	})

	assert.Equal(t, "c", modelsPaging[0].FilterName)
	assert.Equal(t, "d", modelsPaging[1].FilterName)

	// Sorting
	r = httptest.NewRequest(http.MethodGet, "http://abc.de/whatEver?sort=-test", nil)

	urlParams, err = runtime.ReadURLQueryParameters(r, mapper, &testValueSanitizer{})
	require.NoError(t, err)

	var modelsSort []TestModel

	q = db.NewSelect().Model(&modelsSort)
	q = urlParams.AddToQuery(q)

	err = q.Scan(ctx)
	require.NoError(t, err)

	assert.Equal(t, 6, len(modelsSort))
	assert.Equal(t, "f", modelsSort[0].FilterName)
	assert.Equal(t, "e", modelsSort[1].FilterName)
	assert.Equal(t, "d", modelsSort[2].FilterName)
	assert.Equal(t, "c", modelsSort[3].FilterName)
	assert.Equal(t, "b", modelsSort[4].FilterName)
	assert.Equal(t, "a", modelsSort[5].FilterName)

	// Combine all
	r = httptest.NewRequest(http.MethodGet, "http://abc.de/whatEver?sort=-test&filter[test]=a,b,e,f&page[number]=1&page[size]=2", nil)

	urlParams, err = runtime.ReadURLQueryParameters(r, mapper, &testValueSanitizer{})
	require.NoError(t, err)

	var modelsCombined []TestModel

	q = db.NewSelect().Model(&modelsCombined)
	q = urlParams.AddToQuery(q)

	err = q.Scan(ctx)
	require.NoError(t, err)

	assert.Equal(t, 2, len(modelsCombined))
	assert.Equal(t, "b", modelsCombined[0].FilterName)
	assert.Equal(t, "a", modelsCombined[1].FilterName)
}

func setupDatabase(ctx context.Context, t *testing.T) *bun.DB {
	db := postgres.NewDB(context.Background())

	_, err := db.NewCreateTable().Model((*TestModel)(nil)).Exec(ctx)
	require.NoError(t, err)

	testModels := []TestModel{
		{FilterName: "a"},
		{FilterName: "b"},
		{FilterName: "c"},
		{FilterName: "d"},
		{FilterName: "e"},
		{FilterName: "f"},
	}

	_, err = db.NewInsert().Model(&testModels).Exec(ctx)
	require.NoError(t, err)

	return db
}
