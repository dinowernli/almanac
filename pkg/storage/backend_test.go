package storage

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestDiskBackend(t *testing.T) {
	path, err := ioutil.TempDir("", "almanac-test")
	assert.NoError(t, err)
	defer os.RemoveAll(path)

	b := &diskBackend{path}

	// Add some entries.
	assert.NoError(t, b.write(context.Background(), "foo1", []byte("some-content")))
	assert.NoError(t, b.write(context.Background(), "foo2", []byte("some-content")))
	assert.NoError(t, b.write(context.Background(), "foo3", []byte("some-content")))
	assert.NoError(t, b.write(context.Background(), "bar3", []byte("some-content")))

	// Delete one of the foos.
	assert.NoError(t, b.delete(context.Background(), "foo2"))

	// Make sure list works.
	results, err := b.list(context.Background(), "foo")
	assert.NoError(t, err)
	assert.Contains(t, results, "foo1")
	assert.Contains(t, results, "foo3")

	// Read a foo.
	bytes, err := b.read(context.Background(), "bar3")
	assert.NoError(t, err)
	assert.Equal(t, "some-content", string(bytes))
}
