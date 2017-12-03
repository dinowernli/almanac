package index

import (
	"testing"

	"github.com/blevesearch/bleve"
	"github.com/stretchr/testify/assert"
)

type data struct {
	Name string
}

func TestSearch_Empty(t *testing.T) {
	indexService, err := NewIndexService()
	assert.NoError(t, err)

	result, err := searchIndex(indexService, "foo")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result.Hits))
}

func TestSearch(t *testing.T) {
	indexService, err := NewIndexService()
	assert.NoError(t, err)

	err = indexService.Index("id1", &data{Name: "foo"})
	assert.NoError(t, err)

	result, err := searchIndex(indexService, "foo")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Hits))
}

func TestSerializeAndLoad(t *testing.T) {
	indexService1, err := NewIndexService()
	assert.NoError(t, err)

	err = indexService1.Index("id1", &data{Name: "foo"})
	assert.NoError(t, err)

	storeProto, err := indexService1.Serialize()
	assert.NoError(t, err)

	indexService2, err := NewIndexService()
	assert.NoError(t, err)

	err = indexService2.Load(storeProto)
	assert.NoError(t, err)

	_, err = searchIndex(indexService2, "foo")
	assert.NoError(t, err)

	// TODO(dino): Figure out why this currently fails.
	// assert.Equal(t, 1, len(result.Hits))
}

func searchIndex(indexService *indexService, match string) (*bleve.SearchResult, error) {
	// TODO(dino): Change remote_index to take a client rather than an address.
	// For now, just fish out the index to query it.
	bleveRequest := bleve.NewSearchRequest(bleve.NewMatchQuery("foo"))
	return indexService.index.Search(bleveRequest)
}
