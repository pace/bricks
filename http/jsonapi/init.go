// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/05/19 by Vincent Landgraf

package jsonapi

import "github.com/shopspring/decimal"

func init() {
	decimal.MarshalJSONWithoutQuotes = true
}
