package index

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

type data struct {
	Name string
}

func TestSearch_Empty(t *testing.T) {
	index, err := NewIndex()
	assert.NoError(t, err)

	result, err := index.Search(context.Background(), "foo", 200)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
}

func TestSearch(t *testing.T) {
	index, err := NewIndex()
	assert.NoError(t, err)

	err = index.Index("id1", &data{Name: "foo"})
	assert.NoError(t, err)

	result, err := index.Search(context.Background(), "foo", 200)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
}
