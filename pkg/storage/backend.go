package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/net/context"
)

// Storage represents a very basic abstraction which allows reading and writing
// bytes to a persistent medium.
type backend interface {
	// read returns the bytes associated with the given id.
	read(ctx context.Context, id string) ([]byte, error)

	// write stores the supplied bytes under the supplied id.
	write(ctx context.Context, id string, contents []byte) error

	// list returns all keys which start with the supplied prefix.
	list(ctx context.Context, prefix string) ([]string, error)

	// delete removes the bytes associated with the given key.
	delete(ctx context.Context, id string) error
}

// diskBackend is a storage backend backed by a location on disk.
type diskBackend struct {
	path string
}

func (b *diskBackend) read(ctx context.Context, id string) ([]byte, error) {
	return ioutil.ReadFile(b.filename(id))
}

func (b *diskBackend) write(ctx context.Context, id string, contents []byte) error {
	return ioutil.WriteFile(b.filename(id), contents, 0644)
}

func (b *diskBackend) list(ctx context.Context, prefix string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(b.path, prefix+"*"))
	if err != nil {
		return nil, fmt.Errorf("unable to glob files: %v", err)
	}

	results := []string{}
	for _, m := range matches {
		results = append(results, filepath.Base(m))
	}
	return results, nil
}

func (b *diskBackend) delete(ctx context.Context, id string) error {
	filename := filepath.Join(b.path, id)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("value %s does not exist", id)
	}
	err := os.Remove(filename)
	if err != nil {
		return fmt.Errorf("unable to remove file %s: %v", filename, err)
	}
	return nil
}

func (b *diskBackend) filename(id string) string {
	return path.Join(b.path, id)
}

// memoryBackend is a storage backend backed by memory.
type memoryBackend struct {
	data map[string][]byte
}

func (b *memoryBackend) read(ctx context.Context, id string) ([]byte, error) {
	result := b.data[id]
	if result == nil {
		return nil, fmt.Errorf("value %s does not exist", id)
	}
	return result, nil
}

func (b *memoryBackend) write(ctx context.Context, id string, contents []byte) error {
	b.data[id] = contents
	return nil
}

func (b *memoryBackend) list(ctx context.Context, prefix string) ([]string, error) {
	result := []string{}
	for k := range b.data {
		if strings.HasPrefix(k, prefix) {
			result = append(result, k)
		}
	}
	return result, nil
}

func (b *memoryBackend) delete(ctx context.Context, id string) error {
	_, ok := b.data[id]
	if !ok {
		return fmt.Errorf("key %s not found", id)
	}
	delete(b.data, id)
	return nil
}
