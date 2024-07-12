// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package locale

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContext(t *testing.T) {
	_, ok := FromCtx(context.Background())
	assert.False(t, ok)

	l := NewLocale("fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5", "Europe/Paris")
	ctx := WithLocale(context.Background(), l)

	lctx, ok := FromCtx(ctx)
	require.True(t, ok)
	assert.Equal(t, lctx.Language(), "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	assert.Equal(t, lctx.Timezone(), "Europe/Paris")

	trctx := ContextTransfer(ctx, context.Background())

	lctx, ok = FromCtx(trctx)
	require.True(t, ok)
	assert.Equal(t, lctx.Language(), "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	assert.Equal(t, lctx.Timezone(), "Europe/Paris")
}
