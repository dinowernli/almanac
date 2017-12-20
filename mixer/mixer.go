package mixer

import (
	"fmt"

	"dinowernli.me/almanac/discovery"
	"dinowernli.me/almanac/index"
	pb_almanac "dinowernli.me/almanac/proto"
	"dinowernli.me/almanac/storage"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const (
	chunkPrefix = "chunk-"
)

type Mixer struct {
	storage   storage.Storage
	discovery *discovery.Discovery
}

// New returns a new mixer backed by the supplied storage.
func New(storage storage.Storage, discovery *discovery.Discovery) *Mixer {
	return &Mixer{storage: storage, discovery: discovery}
}

func (m *Mixer) Search(ctx context.Context, request *pb_almanac.SearchRequest) (*pb_almanac.SearchResponse, error) {
	indexes := []*index.Index{}

	// Load all relevant chunks as indexes.
	chunks, err := m.loadChunks(request)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "unable to load chunks: %v", err)
	}
	for _, chunkIndex := range chunks {
		indexes = append(indexes, chunkIndex)
	}

	// Gather all relevant appenders.
	appenders, err := m.loadAppenders()
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "unable to load appenders: %v", err)
	}
	for _, appenderIndex := range appenders {
		indexes = append(indexes, appenderIndex)
	}

	// Perform the combined search.
	result, err := index.Search(indexes, ctx, request)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "unable to search indexes: %v", err)
	}

	return result, nil
}

// loadAppenders returns bleve index implementations backed by all appenders in
// the system.
func (m *Mixer) loadAppenders() ([]*index.Index, error) {
	result := []*index.Index{}
	for _, a := range m.discovery.ListAppenders() {
		result = append(result, index.NewRemoteIndex(a))
	}
	return result, nil
}

// loadChunks returns all stored chunks which need to be searched for this request.
func (m *Mixer) loadChunks(request *pb_almanac.SearchRequest) ([]*index.Index, error) {
	// TODO(dino): Stop listing all chunks by using some kind of scan.
	chunkKeys, err := m.storage.List(chunkPrefix)
	if err != nil {
		return nil, fmt.Errorf("unable to list chunk keys: %v", err)
	}

	results := []*index.Index{}
	for _, key := range chunkKeys {
		bytes, err := m.storage.Read(key)
		if err != nil {
			return nil, fmt.Errorf("failed to read key %s: %v", key, err)
		}

		chunk := &pb_almanac.Chunk{}
		err = proto.Unmarshal(bytes, chunk)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal chunk %s: %v", key, err)
		}

		idx, err := index.Deserialize(chunk.Index)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize index from chunk %s: %v", key, err)
		}
		results = append(results, idx)
	}

	return results, nil
}
