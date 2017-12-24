package index

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

type content struct {
	Name string
}

func TestRoundtrip(t *testing.T) {
	index, err := NewIndex()
	assert.NoError(t, err)

	err = index.Index("id1", &content{Name: "foo"})
	assert.NoError(t, err)

	indexProto, err := Serialize(index)
	assert.NoError(t, err)

	deserializedIndex, err := Deserialize(indexProto)
	assert.NoError(t, err)

	result, err := deserializedIndex.Search(context.Background(), "foo", 200)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
}
