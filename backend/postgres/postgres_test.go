// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/12 by Vincent Landgraf

package postgres

import "testing"

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
