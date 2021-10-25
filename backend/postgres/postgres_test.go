// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/12 by Vincent Landgraf

package postgres

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIntegrationConnectionPool(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	db := ConnectionPool()
	var result struct {
		Calc int
	}
	_, err := db.QueryOne(&result, `SELECT ? + ? AS Calc`, 10, 10) //nolint:errcheck
	if err != nil {
		t.Errorf("got %v", err)
	}

	// Note: This test can't actually test the logging correctly
	// but the code will be accessed
}

func TestFirstWords(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	testQuery1 := `SELECT * FROM example`
	require.Equal(t, "SELECT", firstWords(testQuery1, 1))

	testQuery2 := `
		SELECT * FROM example`
	require.Equal(t, "SELECT", firstWords(testQuery2, 1))

	testQuery3 := `
		SELECT * FROM example`
	require.Equal(t, "SELECT *", firstWords(testQuery3, 2))

	testQuery4 := ``
	require.Equal(t, "", firstWords(testQuery4, 1))
}

func TestGetQueryType(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	testQuery1 := `SELECT * FROM example`
	require.Equal(t, "SELECT", getQueryType(testQuery1))

	testQuery2 := `
		SELECT
			* FROM example`
	require.Equal(t, "SELECT", getQueryType(testQuery2))

	testQuery3 := `UPDATE example SET foo = 1`
	require.Equal(t, "UPDATE", getQueryType(testQuery3))

	testQuery4 := `COPY film_locations FROM '/tmp/foo.csv' HEADER CSV DELIMITER ',';`
	require.Equal(t, "CMD", getQueryType(testQuery4))

	badQuery1 := `UPDATEexample SET foo = 1`
	require.Equal(t, "CMD", getQueryType(badQuery1))
}
