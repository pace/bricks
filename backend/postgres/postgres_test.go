// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package postgres

import (
	"testing"

	"github.com/stretchr/testify/require"
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

func TestIntegrationConnectionPoolNoLogging(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	db := ConnectionPool(WithQueryLogging(false, false))
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
	require.Equal(t, "COPY", getQueryType(testQuery4))
}
