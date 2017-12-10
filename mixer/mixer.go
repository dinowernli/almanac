package mixer

import (
	"fmt"

	"dinowernli.me/almanac/discovery"
	"dinowernli.me/almanac/index"
	pb_almanac "dinowernli.me/almanac/proto"
	"dinowernli.me/almanac/storage"

	"github.com/blevesearch/bleve"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const (
	chunkPrefix = "chunk-"
)

type mixer struct {
	storage   storage.Storage
	discovery *discovery.Discovery
}

// New returns a new mixer backed by the supplied storage.
func New(storage storage.Storage, discovery *discovery.Discovery) *mixer {
	return &mixer{storage: storage, discovery: discovery}
}

func (m *mixer) Search(ctx context.Context, request *pb_almanac.SearchRequest) (*pb_almanac.SearchResponse, error) {
	indexes := []bleve.Index{}

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

	// TODO(dino): Create an aliasindex for all the indexes and execute a search.

	return nil, grpc.Errorf(codes.Unimplemented, "search not implemented")
}

func (m *mixer) loadAppenders() ([]bleve.Index, error) {
	result := []bleve.Index{}
	for _, a := range m.discovery.ListAppenders() {
		result = append(result, index.NewRemoteIndex(a))
	}
	return result, nil
}

// loadChunks returns all stored chunks which need to be searched for this request.
func (m *mixer) loadChunks(request *pb_almanac.SearchRequest) ([]bleve.Index, error) {
	// TODO(dino): Add more information to the chunk names, making it easier to only
	// load subsets of all chunks. For now, return them all.
	chunkKeys, err := m.storage.List(chunkPrefix)
	if err != nil {
		return nil, fmt.Errorf("unable to list chunk keys: %v", err)
	}

	result := []*pb_almanac.Chunk{}
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
		result = append(result, chunk)
	}

	// TODO(dino): tranform chunks into indexes.
	return nil, nil
}
