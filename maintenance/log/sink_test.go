package log

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Sink(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	var sink *Sink
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		require.NotEqual(t, "", RequestID(r), "request should have request id")

		var ok bool
		sink, ok = SinkFromContext(r.Context())
		require.True(t, ok, "SinkFromContext() returned false unexpectedly")

		Req(r).Info().Msg("this is a test message for the sink")
		w.WriteHeader(201)
	})
	Handler()(mux).ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	require.Equal(t, 201, resp.StatusCode, "wrong status code")

	logs := sink.ToJSON()

	var result []interface{}
	require.NoError(t, json.Unmarshal(logs, &result), "could not unmarshal logs")

	require.Len(t, result, 2, "expecting exactly one log, but got %d", len(result))
}

func TestOverflowRing(t *testing.T) {
	ring := newStringRing(3)
	for i := 0; i < 2; i++ {
		ring.writeString(fmt.Sprintf("%02d", i))
	}
	require.Equal(t, []string{"00", "01", ""}, ring.data)
	ring.writeString("02")
	require.Equal(t, []string{"00", "01", "02"}, ring.data)
	for i := 3; i < 5; i++ {
		ring.writeString(fmt.Sprintf("%02d", i))
	}
	require.Equal(t, []string{"03", "04", "02"}, ring.data)
}

func TestRingGetContent(t *testing.T) {
	ring := newStringRing(3)
	for i := 0; i < 2; i++ {
		ring.writeString(fmt.Sprintf("%02d", i))
	}
	require.Equal(t, []string{"00", "01"}, ring.GetContent())
	ring.writeString("02")
	require.Equal(t, []string{"00", "01", "02"}, ring.GetContent())
	for i := 3; i < 5; i++ {
		ring.writeString(fmt.Sprintf("%02d", i))
	}
	require.Equal(t, []string{"02", "03", "04"}, ring.GetContent())
}
