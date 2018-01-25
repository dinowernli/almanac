package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimeMs(t *testing.T) {
	assert.Equal(t, int64(1234), TimeMs(1234).UnixNano()/nanosPerMilli)
}
