package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
