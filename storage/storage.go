package storage

import (
	"fmt"
	"io/ioutil"
	"log"
	"path"
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
}

// NewTempDiskStorage creates a storage backed by a new temporary directory.
func NewTempDiskStorage() (Storage, error) {
	path, err := ioutil.TempDir("", tempDirPrefix)
	if err != nil {
		return nil, fmt.Errorf("unable to create temp dir")
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

func (s *diskStorage) filename(id string) string {
	return path.Join(s.path, id)
}
