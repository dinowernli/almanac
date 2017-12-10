package mixer

import (
	"fmt"

	"dinowernli.me/almanac/discovery"
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

type mixer struct {
	storage   storage.Storage
	discovery *discovery.Discovery
}

// New returns a new mixer backed by the supplied storage.
func New(storage storage.Storage, discovery *discovery.Discovery) *mixer {
	return &mixer{storage: storage, discovery: discovery}
}

func (m *mixer) Search(ctx context.Context, request *pb_almanac.SearchRequest) (*pb_almanac.SearchResponse, error) {

	return nil, grpc.Errorf(codes.Unimplemented, "search not implemented")
}

// loadChunks returns all stored chunks which need to be searched for this request.
func (m *mixer) loadChunks(request *pb_almanac.SearchRequest) ([]*pb_almanac.Chunk, error) {
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
	return result, nil
}
