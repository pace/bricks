package isotime

import (
	"testing"
	"time"
)

func TestParseISO8601(t *testing.T) {
	tests := []struct {
		str     string
		want    time.Time
		wantErr bool
	}{
		{
			str:  "1993-12-04T12:42:24.135843287+12:34",
			want: time.Date(1993, 12, 4, 12, 42, 24, 135843287, time.FixedZone("+1234", (12*60*60)+(34*60))),
		},
		{
			str:  "1993-12-04T12:42:24.135843287Z",
			want: time.Date(1993, 12, 4, 12, 42, 24, 135843287, time.UTC),
		},
		{
			str:  "1993-12-04T12:42:24.135843287",
			want: time.Date(1993, 12, 4, 12, 42, 24, 135843287, time.UTC),
		},
		{
			str:  "1993-12-04T12:42:24+12:34",
			want: time.Date(1993, 12, 4, 12, 42, 24, 0, time.FixedZone("+1234", (12*60*60)+(34*60))),
		},
		{
			str:  "1993-12-04T12:42:24Z",
			want: time.Date(1993, 12, 4, 12, 42, 24, 0, time.UTC),
		},
		{
			str:  "1993-12-04T12:42:24",
			want: time.Date(1993, 12, 4, 12, 42, 24, 0, time.UTC),
		},
		{
			str:  "1993-12-04T12:42+12:34",
			want: time.Date(1993, 12, 4, 12, 42, 0, 0, time.FixedZone("+1234", (12*60*60)+(34*60))),
		},
		{
			str:  "1993-12-04T12:42Z",
			want: time.Date(1993, 12, 4, 12, 42, 0, 0, time.UTC),
		},
		{
			str:  "1993-12-04T12:42",
			want: time.Date(1993, 12, 4, 12, 42, 0, 0, time.UTC),
		},
		{
			str:  "1993-12-04",
			want: time.Date(1993, 12, 4, 0, 0, 0, 0, time.UTC),
		},
		{
			str:  "1993-12",
			want: time.Date(1993, 12, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			str:  "1993",
			want: time.Date(1993, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			str:  "",
			want: time.Time{},
		},
		{
			str:     "wqefwefwefwfe",
			wantErr: true,
		},
		{
			str:     "19",
			wantErr: true,
		},
		{
			str:     "19T12",
			wantErr: true,
		},
		{
			str:     "0000-13-85T84:69:88.13584328768468-99:99",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			got, err := ParseISO8601(tt.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseISO8601() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !got.Equal(tt.want) {
				t.Errorf("ParseISO8601() = %v, want %v", got, tt.want)
			}
		})
	}
}
