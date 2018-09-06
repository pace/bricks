// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/09/06 by Vincent Landgraf

package log

import (
	"context"
	"net/http/httptest"
	"testing"
)

func TestLog(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	Req(req).Info().Msg("req")

	Ctx(context.Background()).Info().Msg("ctx")

	Logger().Info().Msg("log")
}
