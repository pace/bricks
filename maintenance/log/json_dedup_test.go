package log

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_JSONDedup_simple(t *testing.T) {
	var buffer bytes.Buffer
	logger := Output(JSONDedup(&buffer)).With().Str("key", "shouldnotbethere").Logger()
	logger.Debug().Str("key", "value").Msg("some message")

	out := buffer.String()
	require.NotContains(t, out, "shouldnotbethere")
	require.Contains(t, out, "value")
	require.Contains(t, out, "some message")
	require.Contains(t, out, "key")
}

func Test_JSONDedup_withSink(t *testing.T) {
	var sink Sink
	ctx := ContextWithSink(WithContext(context.Background()), &sink)

	logger := Ctx(ctx).With().Str("key", "shouldnotbethere").Logger()
	logger.Debug().Str("key", "value").Msg("some message")

	out := string(sink.ToJSON())
	require.NotContains(t, out, "shouldnotbethere")
	require.Contains(t, out, "value")
	require.Contains(t, out, "some message")
	require.Contains(t, out, "key")
}
