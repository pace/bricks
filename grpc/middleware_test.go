// Copyright © 2021  by PACE Telematics GmbH. All rights reserved.

package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	"github.com/pace/bricks/http/middleware"
	"github.com/pace/bricks/pkg/tracking/utm"
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

func TestAddExternalDependencyMetadataToContext(t *testing.T) {
	md := metadata.New(map[string]string{
		MetadataKeyExternalDependencies: "dep1:1,dep2:2,dep3:3",
	})
	ctx := context.Background()
	ctx = AddExternalDependencyMetadataToContext(ctx, md)
	edc := middleware.ExternalDependencyContextFromContext(ctx)
	require.NotNil(t, edc)
	require.Equal(t, "dep1:1,dep2:2,dep3:3", edc.String())

	ctx = AddExternalDependencyMetadataToContext(ctx, metadata.New(map[string]string{
		MetadataKeyExternalDependencies: "dep4:4,dep5:5,dep6:6",
	}))
	edc = middleware.ExternalDependencyContextFromContext(ctx)
	require.NotNil(t, edc)
	require.Equal(t, "dep1:1,dep2:2,dep3:3,dep4:4,dep5:5,dep6:6", edc.String())
}

func TestAddExternalDependencyMetadataToContext_NoDependencies(t *testing.T) {
	md := metadata.New(map[string]string{})
	ctx := context.Background()
	ctx = AddExternalDependencyMetadataToContext(ctx, md)
	edc := middleware.ExternalDependencyContextFromContext(ctx)
	require.Nil(t, edc)
}
