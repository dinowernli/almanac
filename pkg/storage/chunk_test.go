package storage

import (
	"testing"

	pb_almanac "github.com/dinowernli/almanac/proto"

	"github.com/stretchr/testify/assert"
)

var (
	entry1 = &pb_almanac.LogEntry{
		Id:          "id1",
		EntryJson:   `{ "message": "foo" }`,
		TimestampMs: int64(1234),
	}

	entry2 = &pb_almanac.LogEntry{
		Id:          "id2",
		EntryJson:   `{ "message": "foo" }`,
		TimestampMs: int64(5678),
	}
)

func TestParseId(t *testing.T) {
	_, err := ChunkIdProto("sml-123-456-foo")
	assert.NoError(t, err)

	_, err = ChunkIdProto("big-89-25-bar")
	assert.NoError(t, err)
}

func TestRoundtrip(t *testing.T) {
	id := "sml-3-7-asdf"
	idProto, err := ChunkIdProto(id)
	assert.NoError(t, err)

	assert.Equal(t, int64(3), idProto.StartMs)
	assert.Equal(t, int64(7), idProto.EndMs)
	assert.Equal(t, "asdf", idProto.Uid)
	assert.Equal(t, pb_almanac.ChunkId_SMALL, idProto.Type)

	id2, err := ChunkId(idProto)
	assert.NoError(t, err)
	assert.Equal(t, id, id2)
}

func TestChunkProtoCreating(t *testing.T) {
	entriesInput := []*pb_almanac.LogEntry{entry2, entry1}

	chunkProto, err := ChunkProto(entriesInput, pb_almanac.ChunkId_SMALL)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(chunkProto.Entries))
	assert.Equal(t, "id2", chunkProto.Entries[0].Id)
	assert.Equal(t, "id1", chunkProto.Entries[1].Id)
}
