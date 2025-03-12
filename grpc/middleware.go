// Copyright © 2021  by PACE Telematics GmbH. All rights reserved.

package grpc

import (
	"context"
	"encoding/gob"
	"strings"

	"google.golang.org/grpc/metadata"

	"github.com/pace/bricks/http/middleware"
	"github.com/pace/bricks/pkg/tracking/utm"
)

const utmMetadataKey = "utm-bin" // IMPORTANT -bin post-fix allows us to send binary data via grpc metadata, otherwise it will break the protocol

func ContextWithUTMFromMetadata(parentCtx context.Context, md metadata.MD) context.Context {
	dataSlice := md.Get(utmMetadataKey)
	if len(dataSlice) == 0 {
		return parentCtx
	}

	var utmData utm.UTMData
	if err := gob.NewDecoder(strings.NewReader(dataSlice[0])).Decode(&utmData); err != nil {
		return parentCtx
	}

	return utm.ContextWithUTMData(parentCtx, utmData)
}

func EncodeContextWithUTMData(parentCtx context.Context) context.Context {
	utmData, exists := utm.FromContext(parentCtx)
	if !exists {
		return parentCtx
	}

	w := strings.Builder{}
	if err := gob.NewEncoder(&w).Encode(utmData); err != nil {
		return parentCtx
	}

	return metadata.AppendToOutgoingContext(parentCtx, utmMetadataKey, w.String())
}

// AddExternalDependencyMetadataToContext adds external dependencies to the context if they are present in the metadata.
func AddExternalDependencyMetadataToContext(ctx context.Context, md metadata.MD) context.Context {
	// If there are no external dependencies in the metadata, we can return the context as is.
	externalDependencies := md.Get(MetadataKeyExternalDependencies)
	if len(externalDependencies) == 0 {
		return ctx
	}

	// If there are external dependencies in the metadata, we need to parse them and add them to the context.
	edc := middleware.ExternalDependencyContextFromContext(ctx)
	if edc == nil {
		edc = &middleware.ExternalDependencyContext{}
	}

	edc.Parse(externalDependencies[0])

	return middleware.ContextWithExternalDependency(ctx, edc)
}
