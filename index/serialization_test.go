package index

import (
	"testing"

	"github.com/blevesearch/bleve"
	"github.com/stretchr/testify/assert"
)

type content struct {
	Name string
}

func TestRoundtrip(t *testing.T) {
	indexService, err := NewIndexService()
	assert.NoError(t, err)

	err = indexService.Index("id1", &content{Name: "foo"})
	assert.NoError(t, err)

	indexProto, err := Serialize(indexService)
	assert.NoError(t, err)

	deserializedIndex, err := Deserialize(indexProto)
	assert.NoError(t, err)

	result, err := search(deserializedIndex, "foo")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Hits))
}

func search(indexService *indexService, match string) (*bleve.SearchResult, error) {
	return indexService.index.Search(bleve.NewSearchRequest(bleve.NewMatchQuery(match)))
}
