package storage

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	pb_almanac "dinowernli.me/almanac/proto"

	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
)

const (
	tempDirPrefix = "almanac"
	chunkPrefix   = "/chunk/"
)

type storageMetrics struct {
	numLists  prometheus.Counter
	numReads  prometheus.Counter
	numWrites prometheus.Counter
}

// newStorageMetrics returns a struct with metrics registered in the default registry.
func newStorageMetrics() (*storageMetrics, error) {
	result := &storageMetrics{}

	result.numLists = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "almanac_storage_lists",
		Help: "The number of list requests sent to the storage backend",
	})
	if err := prometheus.Register(result.numLists); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			return nil, err
		}
	}

	result.numReads = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "almanac_storage_reads",
		Help: "The number of read requests sent to the storage backend",
	})
	if err := prometheus.Register(result.numReads); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			return nil, err
		}
	}

	result.numWrites = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "almanac_storage_writes",
		Help: "The number of write requests sent to the storage backend",
	})
	if err := prometheus.Register(result.numWrites); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			return nil, err
		}
	}

	return result, nil
}

// Storage stores chunks of log entries in persistent storage and supports
// various ways of loading and storing chunks.
type Storage struct {
	backend backend
	metrics *storageMetrics
}

// ListChunks returns the ids of all stored chunks which overlap with the
// supplied time range (inclusive on both ends).
func (s *Storage) ListChunks(ctx context.Context, startMs int64, endMs int64) ([]string, error) {
	// TODO(dino): Actually respect the start and end times. For now, return all chunks.
	chunkPaths, err := s.backend.list(ctx, chunkPrefix)
	s.metrics.numLists.Inc()
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
	s.metrics.numReads.Inc()
	if err != nil {
		return nil, fmt.Errorf("failed to read chunk %s: %v", chunkId, err)
	}

	chunk := &pb_almanac.Chunk{}
	err = proto.Unmarshal(bytes, chunk)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal chunk %s: %v", chunkId, err)
	}
	return openChunk(chunk)
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
	s.metrics.numWrites.Inc()
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
	return NewDiskStorage(path)
}

// NewDiskStorage creates a backend backed by a root directory on disk.
func NewDiskStorage(path string) (*Storage, error) {
	log.Printf("created storage at path: %s\n", path)
	return newStorage(&diskBackend{path: path})
}

// NewInMemoryStorage returns a storage backed by an in-memory map.
func NewMemoryStorage() (*Storage, error) {
	return newStorage(&memoryBackend{data: map[string][]byte{}})
}

// NewGcsStorage returns a storage backed by the supplied gcs bucket.
func NewGcsStorage(bucketName string) (*Storage, error) {
	backend, err := newGcsBackend(bucketName)
	if err != nil {
		return nil, fmt.Errorf("unable to create gcs backend: %v", err)
	}
	return newStorage(backend)
}

func newStorage(b backend) (*Storage, error) {
	m, err := newStorageMetrics()
	if err != nil {
		return nil, fmt.Errorf("unable to create storage metrics: %v", err)
	}
	return &Storage{metrics: m, backend: b}, nil
}

func chunkKey(chunkId string) string {
	return chunkPrefix + chunkId
}
