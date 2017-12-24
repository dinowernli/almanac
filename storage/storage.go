package storage

import (
	"fmt"
	"io/ioutil"
	"log"

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
	return s.backend.list(chunkPrefix)
}

// LoadChunk loads the chunk with the supplied id. The returned chunk uses
// resources which must be freed once it is no longer in use.
func (s *Storage) LoadChunk(chunkId string) (*Chunk, error) {
	bytes, err := s.backend.read(chunkId)
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
	return s.backend.write(fmt.Sprintf("%s%s", chunkPrefix, chunkId), bytes)
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

// NewInMemoryStorage returns a storage backed by an in-memory map.
func NewInMemoryStorage() *Storage {
	return &Storage{&memoryBackend{data: map[string][]byte{}}}
}
