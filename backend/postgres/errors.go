// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.

package postgres

import (
	"errors"
	"io"
	"net"

	"github.com/uptrace/bun/driver/pgdriver"
)

var ErrNotUnique = errors.New("not unique")

func IsErrConnectionFailed(err error) bool {
	// bun has this check internally for network errors
	if errors.Is(err, io.EOF) {
		return true
	}

	// bun has this check internally for network errors
	_, ok := err.(net.Error)
	if ok {
		return true
	}

	// bun has similar check for integrity violation issues, here we check network issues
	var pgErr pgdriver.Error

	if errors.As(err, &pgErr) {
		code := pgErr.Field('C')
		// We check on error codes of Class 08 — Connection Exception.
		// https://www.postgresql.org/docs/10/errcodes-appendix.html
		if len(code) > 2 && code[0:2] == "08" {
			return true
		}
	}

	return false
}
