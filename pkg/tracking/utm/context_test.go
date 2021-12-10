package utm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextWithUTMData(t *testing.T) {
	ctx := context.Background()
	data := UTMData{
		Source:   "src",
		Medium:   "med",
		Campaign: "camp",
		Term:     "trm",
		Content:  "cnt",
	}
	ctxWithData := ContextWithUTMData(ctx, data)
	_, found := FromContext(ctx)
	assert.False(t, found)
	dataFromCtx, found := FromContext(ctxWithData)
	assert.True(t, found)
	assert.Equal(t, data, dataFromCtx)
}

func TestFromToMap(t *testing.T) {
	data := UTMData{
		Source:   "src",
		Medium:   "med",
		Campaign: "camp",
		Term:     "trm",
		Content:  "cnt",
	}
	m := data.ToMap()
	data2 := FromMap(m)
	require.Equal(t, data, data2)
}
