package appender

import (
	"testing"
	"time"

	"dinowernli.me/almanac/pkg/storage"
	pb_almanac "dinowernli.me/almanac/proto"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

const (
	maxEntries = 3
	maxSpread  = 1500 * time.Millisecond

	// Chosen to be very high so that this doesn't trigger closing for the default
	// test setup. That way, if we fail to close a chunk, the test will time out.
	maxOpenTime = 100 * time.Second
)

var (
	initialEntry = &pb_almanac.LogEntry{
		TimestampMs: 200,
		Id:          "id-initial",
		EntryJson:   `{"message": "foo"}`,
	}

	entry2 = &pb_almanac.LogEntry{
		TimestampMs: 400,
		Id:          "id-2",
		EntryJson:   `{"message": "foo"}`,
	}

	entry3 = &pb_almanac.LogEntry{
		TimestampMs: 600,
		Id:          "id-3",
		EntryJson:   `{"message": "foo"}`,
	}

	entry4 = &pb_almanac.LogEntry{
		TimestampMs: 800000,
		Id:          "id-4",
		EntryJson:   `{"message": "foo"}`,
	}
)

func TestEmptyChunk(t *testing.T) {
	c, _ := newChunk(t)
	c.close()

	chunkProto, err := c.toProto()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(chunkProto.Entries))
	assert.Equal(t, int64(200), chunkProto.Id.StartMs)
	assert.Equal(t, int64(200), chunkProto.Id.EndMs)
	assert.Equal(t, int64(200), chunkProto.Entries[0].TimestampMs)
}

func TestToProto(t *testing.T) {
	c, _ := newChunk(t)
	added, err := c.tryAdd(entry2)
	assert.NoError(t, err)
	assert.True(t, added)

	c.close()

	chunkProto, err := c.toProto()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(chunkProto.Entries))
	assert.Equal(t, int64(200), chunkProto.Id.StartMs)
	assert.Equal(t, int64(400), chunkProto.Id.EndMs)

	// Check that the id is valid.
	_, err = storage.ChunkId(chunkProto.Id)
	assert.NoError(t, err)
}

func TestSearch(t *testing.T) {
	c, _ := newChunk(t)
	added, err := c.tryAdd(entry2)
	assert.NoError(t, err)
	assert.True(t, added)

	results, err := c.search(context.Background(), &pb_almanac.SearchRequest{
		Query:   "foo",
		Num:     200,
		StartMs: 3,
		EndMs:   3000,
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(results))
}

func TestToProtoFailsIfNotClosed(t *testing.T) {
	c, _ := newChunk(t)
	_, err := c.toProto()
	assert.Error(t, err)
}

func TestCloseSendsItselfToSink(t *testing.T) {
	c, sink := newChunk(t)
	c.close()
	out := <-sink
	assert.Equal(t, c, out)
}

func TestClosesAfterMaxEntries(t *testing.T) {
	c, sink := newChunk(t)

	added, err := c.tryAdd(entry2)
	assert.NoError(t, err)
	assert.True(t, added)

	added, err = c.tryAdd(entry3)
	assert.NoError(t, err)
	assert.True(t, added)

	// Make sure that the chunk is closed.
	out := <-sink
	assert.Equal(t, c, out)
}

func TestRejectsEntryDueToSpread(t *testing.T) {
	c, _ := newChunk(t)

	added, err := c.tryAdd(entry4)
	assert.NoError(t, err)
	assert.False(t, added)
}

func TestAutoCloses(t *testing.T) {
	sink := make(chan *openChunk)
	c, err := newOpenChunk(initialEntry, maxEntries, maxSpread, 10 /* maxOpenTimeMs */, sink)
	assert.NoError(t, err)

	// Make sure that the chunk is closed.
	out := <-sink
	assert.Equal(t, c, out)
}

func newChunk(t *testing.T) (*openChunk, chan *openChunk) {
	sink := make(chan *openChunk)
	c, err := newOpenChunk(initialEntry, maxEntries, maxSpread, maxOpenTime, sink)
	assert.NoError(t, err)
	return c, sink
}
