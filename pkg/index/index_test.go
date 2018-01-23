package index

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

type data struct {
	Name string
}

type complexData struct {
	Logger  string   `json:"logger"`
	Message string   `json:"message"`
	Tags    []string `json:"tags"`
}

func TestSearch_Empty(t *testing.T) {
	index, err := NewIndex()
	assert.NoError(t, err)
	defer index.Close()

	result, err := index.Search(context.Background(), "foo", 200)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
}

func TestSearch(t *testing.T) {
	index, err := NewIndex()
	assert.NoError(t, err)
	defer index.Close()

	err = index.Index("id1", &data{Name: "foo"})
	assert.NoError(t, err)

	result, err := index.Search(context.Background(), "foo", 200)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
}

func TestSearch_QueryString(t *testing.T) {
	index, err := NewIndex()
	assert.NoError(t, err)
	defer index.Close()

	err = index.Index("id1", &complexData{
		Logger:  "MyAwesomeLogger",
		Message: "some full text message",
		Tags:    []string{"foo", "bar"},
	})
	assert.NoError(t, err)

	err = index.Index("id2", &complexData{
		Logger:  "MyTerribleLogger",
		Message: "some other full text message",
		Tags:    []string{"bar", "baz"},
	})
	assert.NoError(t, err)

	result, err := index.Search(context.Background(), `logger:NonExistant`, 200)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))

	result, err = index.Search(context.Background(), `logger:MyAwesomeLogger`, 200)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))

	result, err = index.Search(context.Background(), `tags:bar`, 200)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))

	result, err = index.Search(context.Background(), `+tags:bar +logger:MyTerribleLogger`, 200)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
}
