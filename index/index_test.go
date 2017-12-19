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
	indexService, err := NewIndex()
	assert.NoError(t, err)

	result, err := searchIndex(indexService, "foo")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result.Hits))
}

func TestSearch(t *testing.T) {
	indexService, err := NewIndex()
	assert.NoError(t, err)

	err = indexService.Index("id1", &data{Name: "foo"})
	assert.NoError(t, err)

	result, err := searchIndex(indexService, "foo")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Hits))
}

func searchIndex(index *Index, match string) (*bleve.SearchResult, error) {
	// TODO(dino): Change remote_index to take a client rather than an address.
	// For now, just fish out the index to query it.
	bleveRequest := bleve.NewSearchRequest(bleve.NewMatchQuery("foo"))
	return index.index.Search(bleveRequest)
}
