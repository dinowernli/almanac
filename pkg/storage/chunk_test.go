package storage

import (
	"testing"

	pb_almanac "dinowernli.me/almanac/proto"

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

func TestRoundtrip(t *testing.T) {
	id := "3-7-asdf"
	idProto, err := ChunkIdProto(id)
	assert.NoError(t, err)

	assert.Equal(t, int64(3), idProto.StartMs)
	assert.Equal(t, int64(7), idProto.EndMs)
	assert.Equal(t, "asdf", idProto.Uid)

	id2, err := ChunkId(idProto)
	assert.NoError(t, err)
	assert.Equal(t, id, id2)
}

func TestChunkProtoCreating(t *testing.T) {
	entriesInput := []*pb_almanac.LogEntry{entry2, entry1}

	chunkProto, err := ChunkProto(entriesInput)
	assert.NoError(t, err)

	// Check that the entries are sorted in the proto.
	assert.Equal(t, 2, len(chunkProto.Entries))
	assert.Equal(t, "id1", chunkProto.Entries[0].Id)
	assert.Equal(t, "id2", chunkProto.Entries[1].Id)

	// Check that the argument we passed in is unchanged.
	assert.Equal(t, 2, len(entriesInput))
	assert.Equal(t, "id2", entriesInput[0].Id)
	assert.Equal(t, "id1", entriesInput[1].Id)
}
