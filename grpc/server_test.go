// Copyright © 2021 by PACE Telematics GmbH. All rights reserved.

package grpc

import (
	"bytes"
	"context"
	"testing"

	"github.com/pace/bricks/locale"
	"github.com/pace/bricks/maintenance/log"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestPrepareContext(t *testing.T) {
	ctx := context.Background()

	// no data from the remote site
	ctx0, _ := prepareContext(ctx)
	assert.NotEmpty(t, log.RequestIDFromContext(ctx0))

	var buf0 bytes.Buffer
	l := log.Ctx(ctx0).Output(&buf0)
	l.Debug().Msg("test")
	assert.Contains(t, buf0.String(), "{\"level\":\"debug\",\"req_id\":\""+
		log.RequestIDFromContext(ctx0)+"\",\"time\":")
	assert.Contains(t, buf0.String(), ",\"message\":\"test\"}\n")

	// remote site is providing data using a bearer token
	ctx = metadata.NewIncomingContext(ctx, metadata.MD{
		"req_id":       []string{"c690uu0ta2rv348epm8g"},
		"locale":       []string{"fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5|Europe/Paris"},
		"bearer_token": []string{"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"},
	})

	ctx1, md := prepareContext(ctx)
	assert.Len(t, md.Get("req_id"), 0)
	assert.Len(t, md.Get("bearer_token"), 0)
	assert.Equal(t, "c690uu0ta2rv348epm8g", log.RequestIDFromContext(ctx1))
	loc, ok := locale.FromCtx(ctx1)
	assert.True(t, ok)
	assert.Equal(t, "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5", loc.Language())
	assert.Equal(t, "Europe/Paris", loc.Timezone())

	var buf1 bytes.Buffer
	l = log.Ctx(ctx1).Output(&buf1)
	l.Debug().Msg("test")
	assert.Contains(t, buf1.String(), "{\"level\":\"debug\",\"req_id\":\"c690uu0ta2rv348epm8g\",\"time\":\"")
	assert.Contains(t, buf1.String(), ",\"message\":\"test\"}\n")

	// remote site is providing data using a bearer token
	ctx = metadata.NewIncomingContext(ctx, metadata.MD{
		"req_id":       []string{"c690uu0ta2rv348epm8g"},
		"bearer_token": []string{"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"},
	})

	ctx2, md := prepareContext(ctx)
	assert.Len(t, md.Get("req_id"), 0)
	assert.Len(t, md.Get("bearer_token"), 0)
	assert.Equal(t, "c690uu0ta2rv348epm8g", log.RequestIDFromContext(ctx1))

	var buf2 bytes.Buffer
	l = log.Ctx(ctx2).Output(&buf2)
	l.Debug().Msg("test")
	assert.Contains(t, buf2.String(), "{\"level\":\"debug\",\"req_id\":\"c690uu0ta2rv348epm8g\",\"time\":\"")
	assert.Contains(t, buf2.String(), ",\"message\":\"test\"}\n")
	_, ok = locale.FromCtx(ctx2)
	assert.False(t, ok)
}
