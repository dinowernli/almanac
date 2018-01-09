package appender

import (
	"testing"

	pb_almanac "dinowernli.me/almanac/proto"

	"github.com/stretchr/testify/assert"
)

const (
	maxEntries    = 5
	maxSpreadMs   = 500
	maxOpenTimeMs = 500
)

var (
	entry = &pb_almanac.LogEntry{
		TimestampMs: 200,
		Id:          "some-id",
		EntryJson:   "{}",
	}
)

func TestFoo(t *testing.T) {
	sink := make(chan *openChunk)
	c, err := newOpenChunk(entry, maxEntries, maxSpreadMs, maxOpenTimeMs, sink)
	assert.NoError(t, err)

	c.close()
	out := <-sink

	assert.Equal(t, c, out)
}
