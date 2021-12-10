//  Copyright © 2021  by PACE Telematics GmbH. All rights reserved.
//  Created at 2021/12/10  by Julius Foitzik

package grpc

import (
	"context"
	"encoding/gob"
	"strings"

	"github.com/pace/bricks/pkg/tracking/utm"
	"google.golang.org/grpc/metadata"
)

const utmMetadataKey = "utm"

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
