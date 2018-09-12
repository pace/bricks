// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/12 by Vincent Landgraf

package postgres

import "testing"

func TestConnectionPool(t *testing.T) {
	db := ConnectionPool()
	var result struct {
		Calc int
	}
	db.QueryOne(&result, `SELECT ? + ? AS Calc`, 10, 10)

	// Note: This test can't actually test the logging correctly
	// but the code will be accessed
}
