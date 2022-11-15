// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/11/21 by Marius Neugebauer

package postgres

import (
	"errors"
	"io"
	"net"

	"github.com/go-pg/pg"
)

var (
	ErrNotUnique = errors.New("not unique")
)

func IsErrConnectionFailed(err error) bool {
	// go-pg has this check internally for network errors
	if errors.Is(err, io.EOF) {
		return true
	}

	// go-pg has this check internally for network errors
	_, ok := err.(net.Error)
	if ok {
		return true
	}

	// go-pg has similar check for integrity violation issues, here we check network issues
	pgErr, ok := err.(pg.Error)
	if ok {
		code := pgErr.Field('C')
		// We check on error codes of Class 08 — Connection Exception.
		// https://www.postgresql.org/docs/10/errcodes-appendix.html
		if code[0:2] == "08" {
			return true
		}
	}
	return false
}
