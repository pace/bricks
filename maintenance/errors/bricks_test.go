/*
 * Copyright © 2023 by PACE Telematics GmbH. All rights reserved.
 * Created at 2023/1/20 by Sascha Voth
 */

package errors

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBricksError_AsRuntimeError(t *testing.T) {
	want := struct {
		title  string
		detail string
		status int
		code   string
	}{
		title:  "My Error",
		detail: "None",
		status: 666,
		code:   "DEVIL_FOUND",
	}

	e := &BricksError{
		title:  want.title,
		detail: want.detail,
		status: want.status,
		code:   want.code,
	}
	r := e.AsRuntimeError()

	assert.Equalf(t, want.code, r.Code, "AsRuntimeError().Code")
	assert.Equalf(t, want.title, r.Title, "AsRuntimeError().Title")
	assert.Equalf(t, strconv.Itoa(want.status), r.Status, "AsRuntimeError().Status")
	assert.Equalf(t, want.detail, r.Detail, "AsRuntimeError().Detail")
}

func TestBricksError_Options(t *testing.T) {
	assert.Equalf(t, "DETAIL", NewBricksError(WithDetail("DETAIL")).detail, "WithDetail()")
	assert.Equalf(t, 333, NewBricksError(WithStatus(333)).status, "WithStatus()")
	assert.Equalf(t, "TITLE", NewBricksError(WithTitle("TITLE")).title, "WithTitle()")
	assert.Equalf(t, "CODE", NewBricksError(WithCode("CODE")).code, "WithCode()")
}
