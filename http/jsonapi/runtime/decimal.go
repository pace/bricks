package runtime

import "github.com/shopspring/decimal"

type Decimal string

func DecimalFrom(d decimal.Decimal) Decimal {
	return Decimal(d.String())
}

func (d Decimal) Decode() decimal.Decimal {
	return decimal.RequireFromString(string(d))
}
