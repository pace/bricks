// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/10/31 by Charlotte Pröller

package security

import (
	"context"
	"net/http"
)

type Authorizer interface {
	Authorize(cfg interface{}, r *http.Request, w http.ResponseWriter) (context.Context, bool)
}
