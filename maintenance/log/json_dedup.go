package log

import (
	"bytes"
	"encoding/json"
	"io"
)

type jsonDedup struct {
	output io.Writer
}

func JSONDedup(output io.Writer) io.Writer {
	return &jsonDedup{output: output}
}

func (j *jsonDedup) Write(p []byte) (n int, err error) {
	var payload map[string]interface{}
	if err := json.NewDecoder(bytes.NewReader(p)).Decode(&payload); err != nil {
		return 0, err
	}

	return len(p), json.NewEncoder(j.output).Encode(payload)
}
