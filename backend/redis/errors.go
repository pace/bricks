// Copyright © 2022 by PACE Telematics GmbH. All rights reserved.

package redis

import (
	"errors"
	"io"
	"net"
)

func IsErrConnectionFailed(err error) bool {
	// go-redis has this check internally for network errors
	if errors.Is(err, io.EOF) {
		return true
	}

	// go-redis has this check internally for network errors
	_, ok := err.(net.Error)
	return ok
}
