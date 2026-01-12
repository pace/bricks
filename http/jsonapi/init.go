// Copyright © 2020–2026 by PACE Mobility GmbH. All rights reserved.

package jsonapi

import "github.com/shopspring/decimal"

func init() {
	decimal.MarshalJSONWithoutQuotes = true
}
