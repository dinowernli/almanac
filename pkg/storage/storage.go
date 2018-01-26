package storage

import (
	"fmt"
	"io"
	"os"
	"strings"

	"dinowernli.me/almanac/pkg/util"
	pb_almanac "dinowernli.me/almanac/proto"

	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
)

const (
	StorageTypeMemory = "memory"
	StorageTypeDisk   = "disk"
	StorageTypeGcs    = "gcs"

	chunkPrefix    = "chunk-"
	chunkTypeLabel = "chunk_type"
)

type storageMetrics struct {
	numLists   *prometheus.CounterVec
	numReads   prometheus.Counter
	numWrites  prometheus.Counter
	numDeletes prometheus.Counter
}

// newStorageMetrics returns a struct with metrics registered in the default registry.
func newStorageMetrics() (*storageMetrics, error) {
	result := &storageMetrics{}

	result.numLists = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "almanac_storage_lists",
		Help: "The number of list requests sent to the storage backend",
	}, []string{chunkTypeLabel})
	if err := util.RegisterLenient(result.numLists); err != nil {
		return nil, err
	}

	result.numReads = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "almanac_storage_reads",
		Help: "The number of read requests sent to the storage backend",
	})
	if err := util.RegisterLenient(result.numReads); err != nil {
		return nil, err
	}

	result.numWrites = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "almanac_storage_writes",
		Help: "The number of write requests sent to the storage backend",
	})
	if err := util.RegisterLenient(result.numWrites); err != nil {
		return nil, err
	}

	result.numDeletes = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "almanac_storage_deletes",
		Help: "The number of delete requests sent to the storage backend",
	})
	if err := util.RegisterLenient(result.numDeletes); err != nil {
		return nil, err
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
func (s *Storage) ListChunks(ctx context.Context, startMs int64, endMs int64, chunkType pb_almanac.ChunkId_Type) ([]string, error) {
	chunkTypeString, ok := chunkTypeString[chunkType]
	if !ok {
		return nil, fmt.Errorf("unknown chunk type: %v", chunkType)
	}

	// TODO(dino): Actually respect the start and end times. For now, return all chunks.
	chunkPaths, err := s.backend.list(ctx, chunkPrefix+chunkTypeString)
	s.metrics.numLists.With(prometheus.Labels{chunkTypeLabel: chunkTypeString})
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
func (s *Storage) LoadChunk(ctx context.Context, chunkIdProto *pb_almanac.ChunkId) (*Chunk, error) {
	chunkId, err := ChunkId(chunkIdProto)
	if err != nil {
		return nil, fmt.Errorf("unable to compute chunk id from proto: %v", err)
	}

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

func (s *Storage) DeleteChunk(ctx context.Context, chunkIdProto *pb_almanac.ChunkId) error {
	chunkId, err := ChunkId(chunkIdProto)
	if err != nil {
		return fmt.Errorf("unable to extract chunk id: %v", err)
	}
	err = s.backend.delete(ctx, chunkKey(chunkId))
	s.metrics.numDeletes.Inc()
	if err != nil {
		return fmt.Errorf("unable to delete chunk: %v", err)
	}
	return nil
}

// NewDiskStorage creates a backend backed by a root directory on disk. The supplied path
// must point to an existing empty directory.
func NewDiskStorage(path string) (*Storage, error) {
	empty, err := isEmptyDir(path)
	if err != nil {
		return nil, fmt.Errorf("unable to determine whether path %s is empty dir: %v", path, err)
	}
	if !empty {
		return nil, fmt.Errorf("expected path %s to be an empty directory, but wasn't", path)
	}
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

func isEmptyDir(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	} else {
		return false, err
	}

	return false, nil
}
