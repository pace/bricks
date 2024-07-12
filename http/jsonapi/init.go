// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package jsonapi

import "github.com/shopspring/decimal"

func init() {
	decimal.MarshalJSONWithoutQuotes = true
}
