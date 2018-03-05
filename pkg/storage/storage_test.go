package storage

import (
	"testing"

	pb_almanac "github.com/dinowernli/almanac/proto"

	"reflect"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

var (
	entry = &pb_almanac.LogEntry{
		Id:          "some-id",
		EntryJson:   `{ "message": "foo" }`,
		TimestampMs: int64(1234),
	}
)

func TestStorageRoundTrip(t *testing.T) {
	chunkProto, err := ChunkProto([]*pb_almanac.LogEntry{entry}, pb_almanac.ChunkId_SMALL)
	assert.NoError(t, err)

	chunk, err := openChunk(chunkProto)
	assert.NoError(t, err)
	defer chunk.Close()

	storage, err := NewMemoryStorage()
	assert.NoError(t, err)

	_, err = storage.StoreChunk(context.Background(), chunkProto)
	assert.NoError(t, err)

	loadedChunk, err := storage.LoadChunk(context.Background(), chunkProto.Id)
	assert.NoError(t, err)
	defer loadedChunk.Close()

	assert.True(t, reflect.DeepEqual(chunk.entryMap, loadedChunk.entryMap))
}

func TestList(t *testing.T) {
	storage, err := NewMemoryStorage()
	assert.NoError(t, err)

	chunkProto, err := ChunkProto([]*pb_almanac.LogEntry{entry}, pb_almanac.ChunkId_SMALL)
	assert.NoError(t, err)

	_, err = storage.StoreChunk(context.Background(), chunkProto)
	assert.NoError(t, err)

	smallChunks, err := storage.ListChunks(context.Background(), 0, 0, pb_almanac.ChunkId_SMALL)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(smallChunks))

	bigChunks, err := storage.ListChunks(context.Background(), 0, 0, pb_almanac.ChunkId_BIG)
	assert.NoError(t, err)
	assert.Empty(t, bigChunks)
}

func TestDelete(t *testing.T) {
	storage, err := NewMemoryStorage()
	assert.NoError(t, err)

	chunkProto, err := ChunkProto([]*pb_almanac.LogEntry{entry}, pb_almanac.ChunkId_SMALL)
	assert.NoError(t, err)

	// Try deleting before it's present.
	err = storage.DeleteChunk(context.Background(), chunkProto.Id)
	assert.Error(t, err)

	_, err = storage.StoreChunk(context.Background(), chunkProto)
	assert.NoError(t, err)

	smallChunks, err := storage.ListChunks(context.Background(), 0, 0, pb_almanac.ChunkId_SMALL)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(smallChunks))

	// Try deleting again.
	err = storage.DeleteChunk(context.Background(), chunkProto.Id)
	assert.NoError(t, err)

	smallChunks, err = storage.ListChunks(context.Background(), 0, 0, pb_almanac.ChunkId_SMALL)
	assert.NoError(t, err)
	assert.Empty(t, smallChunks)
}
