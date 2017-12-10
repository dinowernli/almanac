package storage

import (
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"strings"
)

const (
	tempDirPrefix = "almanac"
)

// Storage represents a very basic abstraction which allows reading and writing
// bytes to a persistent medium.
type Storage interface {
	// Read returns the bytes associated with the given id.
	Read(id string) ([]byte, error)

	// Write stores the supplied bytes under the supplied id.
	Write(id string, contents []byte) error

	// List returns all keys which start with the supplied prefix.
	List(prefix string) ([]string, error)
}

// NewTempDiskStorage creates a storage backed by a new temporary directory.
func NewTempDiskStorage() (Storage, error) {
	path, err := ioutil.TempDir("", tempDirPrefix)
	if err != nil {
		return nil, fmt.Errorf("unable to create temp dir: %v")
	}
	return NewDiskStorage(path), nil
}

// NewDiskStorage creates a storage backed by a root directory on disk.
func NewDiskStorage(path string) Storage {
	log.Printf("created storage at path: %s\n", path)
	return &diskStorage{path: path}
}

type diskStorage struct {
	path string
}

func (s *diskStorage) Read(id string) ([]byte, error) {
	return ioutil.ReadFile(s.filename(id))
}

func (s *diskStorage) Write(id string, contents []byte) error {
	return ioutil.WriteFile(s.filename(id), contents, 0644)
}

func (s *diskStorage) List(prefix string) ([]string, error) {
	return nil, fmt.Errorf("List is not implemented for disk storage")
}

func (s *diskStorage) filename(id string) string {
	return path.Join(s.path, id)
}

// NewInMemoryStorage returns a storage backed by an in-memory map.
func NewInMemoryStorage() Storage {
	return &inMemoryStorage{data: map[string][]byte{}}
}

type inMemoryStorage struct {
	data map[string][]byte
}

func (s *inMemoryStorage) Read(id string) ([]byte, error) {
	result := s.data[id]
	if result == nil {
		return nil, fmt.Errorf("value %s does not exist", id)
	} else {
		return result, nil
	}
}

func (s *inMemoryStorage) Write(id string, contents []byte) error {
	s.data[id] = contents
	return nil
}

func (s *inMemoryStorage) List(prefix string) ([]string, error) {
	result := []string{}
	for k, _ := range s.data {
		if strings.HasPrefix(k, prefix) {
			result = append(result, k)
		}
	}
	return result, nil
}
