package storage

import (
	"fmt"
	"io/ioutil"
	"path"
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
}

// diskBackend is a storage backend backed by a location on disk.
type diskBackend struct {
	path string
}

func (s *diskBackend) read(ctx context.Context, id string) ([]byte, error) {
	return ioutil.ReadFile(s.filename(id))
}

func (s *diskBackend) write(ctx context.Context, id string, contents []byte) error {
	return ioutil.WriteFile(s.filename(id), contents, 0644)
}

func (s *diskBackend) list(ctx context.Context, prefix string) ([]string, error) {
	return nil, fmt.Errorf("List is not implemented for disk storage")
}

func (s *diskBackend) filename(id string) string {
	return path.Join(s.path, id)
}

// memoryBackend is a storage backend backed by memory.
type memoryBackend struct {
	data map[string][]byte
}

func (s *memoryBackend) read(ctx context.Context, id string) ([]byte, error) {
	result := s.data[id]
	if result == nil {
		return nil, fmt.Errorf("value %s does not exist", id)
	}
	return result, nil
}

func (s *memoryBackend) write(ctx context.Context, id string, contents []byte) error {
	s.data[id] = contents
	return nil
}

func (s *memoryBackend) list(ctx context.Context, prefix string) ([]string, error) {
	result := []string{}
	for k := range s.data {
		if strings.HasPrefix(k, prefix) {
			result = append(result, k)
		}
	}
	return result, nil
}
