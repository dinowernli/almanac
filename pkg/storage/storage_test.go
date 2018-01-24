package storage

import (
	"testing"

	pb_almanac "dinowernli.me/almanac/proto"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"reflect"
)

var (
	entry = &pb_almanac.LogEntry{
		Id:          "some-id",
		EntryJson:   `{ "message": "foo" }`,
		TimestampMs: int64(1234),
	}
)

func TestStorageRoundTrip(t *testing.T) {
	chunkProto, err := ChunkProto([]*pb_almanac.LogEntry{entry})
	assert.NoError(t, err)

	chunk, err := openChunk(chunkProto)
	assert.NoError(t, err)
	defer chunk.Close()

	storage, err := NewMemoryStorage()
	assert.NoError(t, err)

	id, err := storage.StoreChunk(context.Background(), chunkProto)
	assert.NoError(t, err)

	loadedChunk, err := storage.LoadChunk(context.Background(), id)
	assert.NoError(t, err)
	defer loadedChunk.Close()

	assert.True(t, reflect.DeepEqual(chunk.entries, loadedChunk.entries))
}
