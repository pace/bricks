// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/02/06 by Charlotte Pröller

package runtime_test

import (
	"net/http/httptest"
	"testing"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/pace/bricks/backend/postgres"
	"github.com/pace/bricks/http/jsonapi/runtime"
	"github.com/stretchr/testify/require"
)

type TestModel struct {
	id         int
	filterName string
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
	db := setupDatabase(t)
	r := httptest.NewRequest("GET", "http://abc.de/whatEver?filter[test]=b", nil)
	mappingNames := map[string]string{
		"test": "filterName",
	}
	filterFunc, err := runtime.FilterFromRequest(r, mappingNames, &testValueSanitizer{})
	_ = filterFunc
	require.NoError(t, err)
	var models []TestModel
	q := db.Model(&models)
	//q, err = filterFunc(q)
	require.NoError(t, err)
	count, _ := q.SelectAndCount()
	require.Equal(t, 1, count)
	require.Equal(t, 1, models[0].id)
	require.Equal(t, "b", models[0].filterName)
}

func setupDatabase(t *testing.T) *pg.DB {
	dB := postgres.DefaultConnectionPool()

	require.NoError(t, dB.CreateTable(&TestModel{}, &orm.CreateTableOptions{Temp: true}))
	require.NoError(t, dB.Insert(&TestModel{
		id:         0,
		filterName: "a",
	}))
	require.NoError(t, dB.Insert(&TestModel{
		id:         1,
		filterName: "b",
	}))
	require.NoError(t, dB.Insert(&TestModel{
		id:         2,
		filterName: "c",
	}))
	return dB
}
