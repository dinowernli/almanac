package janitor

import (
	"testing"
	"time"

	st "dinowernli.me/almanac/pkg/storage"
	pb_almanac "dinowernli.me/almanac/proto"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

const (
	compactionInterval = 1 * time.Second
	bigChunkMaxSpread  = 10 * time.Millisecond
)

var (
	entry1 = &pb_almanac.LogEntry{Id: "id1", TimestampMs: 1, EntryJson: `{}`}
	entry2 = &pb_almanac.LogEntry{Id: "id2", TimestampMs: 2, EntryJson: `{}`}
	entry3 = &pb_almanac.LogEntry{Id: "id3", TimestampMs: 3, EntryJson: `{}`}
	entry4 = &pb_almanac.LogEntry{Id: "id4", TimestampMs: 4, EntryJson: `{}`}
	entry5 = &pb_almanac.LogEntry{Id: "id4", TimestampMs: 40000, EntryJson: `{}`}
	entry6 = &pb_almanac.LogEntry{Id: "id4", TimestampMs: 50000, EntryJson: `{}`}
)

func TestCompaction(t *testing.T) {
	storage := createStorage(t)

	// Make sure there are no big chunks to start with.
	bigChunks, err := storage.ListChunks(context.Background(), 0, 0, pb_almanac.ChunkId_BIG)
	assert.NoError(t, err)
	assert.Empty(t, bigChunks)

	_, err = New(context.Background(), logrus.New(), storage, compactionInterval, bigChunkMaxSpread)
	assert.NoError(t, err)

	// Give the janitor enough time to compact.
	time.Sleep(4 * compactionInterval)

	// Check that we have big chunks now.
	bigChunks, err = storage.ListChunks(context.Background(), 0, 0, pb_almanac.ChunkId_BIG)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(bigChunks))

	bigChunkIdProto, err := st.ChunkIdProto(bigChunks[0])
	assert.NoError(t, err)

	bigChunk, err := storage.LoadChunk(context.Background(), bigChunkIdProto)
	assert.NoError(t, err)
	defer bigChunk.Close()

	assert.Equal(t, int64(1), bigChunk.Id().StartMs)
	assert.Equal(t, int64(4), bigChunk.Id().EndMs)
}

func createStorage(t *testing.T) *st.Storage {
	storage, err := st.NewMemoryStorage()
	assert.NoError(t, err)

	chunk1, err := st.ChunkProto([]*pb_almanac.LogEntry{entry1, entry2}, pb_almanac.ChunkId_SMALL)
	assert.NoError(t, err)
	_, err = storage.StoreChunk(context.Background(), chunk1)
	assert.NoError(t, err)

	chunk2, err := st.ChunkProto([]*pb_almanac.LogEntry{entry3, entry4}, pb_almanac.ChunkId_SMALL)
	assert.NoError(t, err)
	_, err = storage.StoreChunk(context.Background(), chunk2)
	assert.NoError(t, err)

	chunk3, err := st.ChunkProto([]*pb_almanac.LogEntry{entry5, entry6}, pb_almanac.ChunkId_SMALL)
	assert.NoError(t, err)
	_, err = storage.StoreChunk(context.Background(), chunk3)
	assert.NoError(t, err)

	return storage
}
