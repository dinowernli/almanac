package storage

import (
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"strings"

	pb_almanac "dinowernli.me/almanac/proto"

	"github.com/golang/protobuf/proto"
)

const (
	tempDirPrefix = "almanac"
	chunkPrefix   = "/chunk/"
)

// Storage stores chunks of log entries in persistent storage and supports
// various ways of loading and storing chunks.
type Storage struct {
	backend backend
}

// ListChunks returns the ids of all stored chunks which overlap with the
// supplied time range (inclusive on both ends).
func (s *Storage) ListChunks(startMs int64, endMs int64) ([]string, error) {
	// TODO(dino): Actually respect the start and end times. For now, return
	// all chunks.
	return s.backend.List(chunkPrefix)
}

// LoadChunk loads the chunk with the supplied id. The returned chunk uses
// resources which must be freed once it is no longer in use.
func (s *Storage) LoadChunk(chunkId string) (*Chunk, error) {
	bytes, err := s.backend.Read(chunkId)
	if err != nil {
		return nil, fmt.Errorf("failed to read key %s: %v", chunkId, err)
	}

	chunk := &pb_almanac.Chunk{}
	err = proto.Unmarshal(bytes, chunk)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal chunk %s: %v", chunkId, err)
	}
	return openChunk(chunkId, chunk)
}

// StoreChunk persists the supplied chunk proto in storage.
func (s *Storage) StoreChunk(chunkId string, chunkProto *pb_almanac.Chunk) error {
	bytes, err := proto.Marshal(chunkProto)
	if err != nil {
		return fmt.Errorf("unable to marshal chunk proto: %v", err)
	}
	return s.backend.Write(fmt.Sprintf("%s%s", chunkPrefix, chunkId), bytes)
}

// Storage represents a very basic abstraction which allows reading and writing
// bytes to a persistent medium.
type backend interface {
	// Read returns the bytes associated with the given id.
	Read(id string) ([]byte, error)

	// Write stores the supplied bytes under the supplied id.
	Write(id string, contents []byte) error

	// List returns all keys which start with the supplied prefix.
	List(prefix string) ([]string, error)
}

// NewTempDiskStorage creates a backend backed by a new temporary directory.
func NewTempDiskStorage() (*Storage, error) {
	path, err := ioutil.TempDir("", tempDirPrefix)
	if err != nil {
		return nil, fmt.Errorf("unable to create temp dir: %v")
	}
	return NewDiskStorage(path), nil
}

// NewDiskStorage creates a backend backed by a root directory on disk.
func NewDiskStorage(path string) *Storage {
	log.Printf("created storage at path: %s\n", path)
	return &Storage{&diskBackend{path: path}}
}

type diskBackend struct {
	path string
}

func (s *diskBackend) Read(id string) ([]byte, error) {
	return ioutil.ReadFile(s.filename(id))
}

func (s *diskBackend) Write(id string, contents []byte) error {
	return ioutil.WriteFile(s.filename(id), contents, 0644)
}

func (s *diskBackend) List(prefix string) ([]string, error) {
	return nil, fmt.Errorf("List is not implemented for disk storage")
}

func (s *diskBackend) filename(id string) string {
	return path.Join(s.path, id)
}

// NewInMemoryStorage returns a storage backed by an in-memory map.
func NewInMemoryStorage() *Storage {
	return &Storage{&memoryBackend{data: map[string][]byte{}}}
}

type memoryBackend struct {
	data map[string][]byte
}

func (s *memoryBackend) Read(id string) ([]byte, error) {
	result := s.data[id]
	if result == nil {
		return nil, fmt.Errorf("value %s does not exist", id)
	} else {
		return result, nil
	}
}

func (s *memoryBackend) Write(id string, contents []byte) error {
	s.data[id] = contents
	return nil
}

func (s *memoryBackend) List(prefix string) ([]string, error) {
	result := []string{}
	for k, _ := range s.data {
		if strings.HasPrefix(k, prefix) {
			result = append(result, k)
		}
	}
	return result, nil
}
