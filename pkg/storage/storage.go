package storage

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	pb_almanac "dinowernli.me/almanac/proto"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
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
func (s *Storage) ListChunks(ctx context.Context, startMs int64, endMs int64) ([]string, error) {
	// TODO(dino): Actually respect the start and end times. For now, return all chunks.

	chunkPaths, err := s.backend.list(ctx, chunkPrefix)
	if err != nil {
		return nil, fmt.Errorf("unable to list chunks: %v", err)
	}
	results := []string{}
	for _, path := range chunkPaths {
		results = append(results, strings.TrimPrefix(path, chunkPrefix))
	}
	return results, nil
}

// LoadChunk loads the chunk with the supplied id. The returned chunk uses
// resources which must be freed once it is no longer in use.
func (s *Storage) LoadChunk(ctx context.Context, chunkId string) (*Chunk, error) {
	bytes, err := s.backend.read(ctx, chunkKey(chunkId))
	if err != nil {
		return nil, fmt.Errorf("failed to read chunk %s: %v", chunkId, err)
	}

	chunk := &pb_almanac.Chunk{}
	err = proto.Unmarshal(bytes, chunk)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal chunk %s: %v", chunkId, err)
	}
	return openChunk(chunkId, chunk)
}

// StoreChunk persists the supplied chunk proto in storage. Returns the id used
// to store the chunk.
func (s *Storage) StoreChunk(ctx context.Context, chunkProto *pb_almanac.Chunk) (string, error) {
	chunkId, err := ChunkId(chunkProto.Id)
	if err != nil {
		return "", fmt.Errorf("unable to extract chunk id: %v", err)
	}

	bytes, err := proto.Marshal(chunkProto)
	if err != nil {
		return "", fmt.Errorf("unable to marshal chunk proto: %v", err)
	}

	err = s.backend.write(ctx, chunkKey(chunkId), bytes)
	if err != nil {
		return "", fmt.Errorf("unable to write chunk bytes to backend: %v", err)
	}

	return chunkId, nil
}

// NewTempDiskStorage creates a backend backed by a new temporary directory.
func NewTempDiskStorage() (*Storage, error) {
	path, err := ioutil.TempDir("", tempDirPrefix)
	if err != nil {
		return nil, fmt.Errorf("unable to create temp dir: %v", err)
	}
	return NewDiskStorage(path), nil
}

// NewDiskStorage creates a backend backed by a root directory on disk.
func NewDiskStorage(path string) *Storage {
	log.Printf("created storage at path: %s\n", path)
	return &Storage{&diskBackend{path: path}}
}

// NewInMemoryStorage returns a storage backed by an in-memory map.
func NewMemoryStorage() *Storage {
	return &Storage{&memoryBackend{data: map[string][]byte{}}}
}

// NewGcsStorage returns a storage backed by the supplied gcs bucket.
func NewGcsStorage(bucketName string) (*Storage, error) {
	backend, err := newGcsBackend(bucketName)
	if err != nil {
		return nil, fmt.Errorf("unable to create gcs backend: %v", err)
	}
	return &Storage{backend}, nil
}

func chunkKey(chunkId string) string {
	return chunkPrefix + chunkId
}
