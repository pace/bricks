//  Copyright © 2022  by PACE Telematics GmbH. All rights reserved.
//  Created at 2022/1/13  by Julius Foitzik

package hlog

import (
	"context"
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
)

func TestInOutContextEmpty(t *testing.T) {
	ctx := context.Background()
	_, found := IDFromCtx(ctx)
	require.False(t, found)
}

func TestInOutContext(t *testing.T) {
	newID := xid.New()
	ctx := WithValue(context.Background(), newID)
	id, found := IDFromCtx(ctx)
	require.True(t, found)
	require.Equal(t, newID.Bytes(), id.Bytes())
}
