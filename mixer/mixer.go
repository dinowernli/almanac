package mixer

import (
	"fmt"

	"dinowernli.me/almanac/discovery"
	"dinowernli.me/almanac/index"
	pb_almanac "dinowernli.me/almanac/proto"
	"dinowernli.me/almanac/storage"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type Mixer struct {
	storage   *storage.Storage
	discovery *discovery.Discovery
}

// New returns a new mixer backed by the supplied storage.
func New(storage *storage.Storage, discovery *discovery.Discovery) *Mixer {
	return &Mixer{storage: storage, discovery: discovery}
}

func (m *Mixer) Search(ctx context.Context, request *pb_almanac.SearchRequest) (*pb_almanac.SearchResponse, error) {
	indexes := []*index.Index{}

	// Load all relevant chunks as indexes.
	chunks, err := m.loadChunks(request)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "unable to load chunks: %v", err)
	}
	defer func() {
		for _, chunk := range chunks {
			chunk.Close()
		}
	}()

	for _, chunk := range chunks {
		indexes = append(indexes, chunk.Index())
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
func (m *Mixer) loadChunks(request *pb_almanac.SearchRequest) ([]*storage.Chunk, error) {
	chunkIds, err := m.storage.ListChunks(request.StartMs, request.EndMs)
	if err != nil {
		return nil, fmt.Errorf("unable to list chunks: %v", err)
	}

	results := []*storage.Chunk{}
	for _, chunkId := range chunkIds {
		chunk, err := m.storage.LoadChunk(chunkId)
		if err != nil {
			return nil, fmt.Errorf("unable to load chunk %s: %v", chunkId, err)
		}
		results = append(results, chunk)
	}
	return results, nil
}
