package runtime

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestDecimal_Decode(t *testing.T) {
	tests := []struct {
		name      string
		d         string
		want      decimal.Decimal
		wantPanic bool
	}{
		{
			name: "valid_decimal",
			d:    "12.34567",
			want: decimal.RequireFromString("12.34567"),
		},
		{
			name: "empty_string",
			d:    "",
			want: decimal.Zero,
		},
		{
			name:      "invalid_decimal",
			d:         "thisisaninvaliddecimal123!!!",
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := func() {
				if got := Decimal(tt.d).Decode(); !got.Equals(tt.want) {
					t.Errorf("Decimal.Decode() = %v, want %v", got, tt.want)
				}
			}

			if tt.wantPanic {
				assert.Panics(t, assert.PanicTestFunc(f))
			} else {
				f()
			}
		})
	}
}
