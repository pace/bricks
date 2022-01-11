//  Copyright © 2021  by PACE Telematics GmbH. All rights reserved.
//  Created at 2021/12/10  by Julius Foitzik

package grpc

import (
	"context"
	"testing"

	"github.com/pace/bricks/pkg/tracking/utm"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestEncodeContextWithUTMData(t *testing.T) {
	data := utm.UTMData{
		Source:   "src",
		Medium:   "med",
		Campaign: "camp",
		Term:     "trm",
		Content:  "cnt",
	}
	ctx := context.Background()
	ctx = utm.ContextWithUTMData(ctx, data)
	ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{}))
	ctx = EncodeContextWithUTMData(ctx)
	md, exists := metadata.FromOutgoingContext(ctx)
	require.True(t, exists)
	ctx2 := context.Background()
	ctx2 = ContextWithUTMFromMetadata(ctx2, md)
	utmData, exists := utm.FromContext(ctx2)
	require.True(t, exists)
	require.Equal(t, data, utmData)
}
