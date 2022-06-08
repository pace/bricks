package terminationlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrimLogFileOutput(t *testing.T) {
	b := make([]byte, 5000)
	assert.Len(t, limitLogFileOutput(string(b)), 4096)
	b = make([]byte, 2000)
	assert.Len(t, limitLogFileOutput(string(b)), 2000)
	b = make([]byte, 4096)
	assert.Len(t, limitLogFileOutput(string(b)), 4096)
}
