package isotime

import (
	"fmt"
	"time"
)

var iso8601Layouts = []string{
	"2006-01-02T15:04:05.999999999Z07:00",
	"2006-01-02T15:04:05.999999999",
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02T15:04:05",
	"2006-01-02T15:04Z07:00",
	"2006-01-02T15:04",
	"2006-01-02",
	"2006-01",
	"2006",
}

func ParseISO8601(str string) (time.Time, error) {
	if str == "" {
		return time.Time{}, nil
	}

	var t time.Time

	var err error

	for _, l := range iso8601Layouts {
		t, err = time.Parse(l, str)
		if err == nil {
			return t, nil
		}
	}

	return t, fmt.Errorf("%q is not a valid ISO8601: %w", str, err)
}
