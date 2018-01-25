package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTimeMs(t *testing.T) {
	assert.Equal(t, int64(1234), TimeMs(1234).UnixNano()/nanosPerMilli)
}
