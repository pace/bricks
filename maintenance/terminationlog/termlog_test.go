package terminationlog

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTrimLogFileOutput(t *testing.T) {
	b := make([]byte, 5000)
	assert.Len(t, limitLogFileOutput(string(b)), 4096)
	b = make([]byte, 2000)
	assert.Len(t, limitLogFileOutput(string(b)), 2000)
	b = make([]byte, 4096)
	assert.Len(t, limitLogFileOutput(string(b)), 4096)
}
