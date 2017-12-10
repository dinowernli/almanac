package mixer

import (
	"dinowernli.me/almanac/discovery"
	pb_almanac "dinowernli.me/almanac/proto"
	"dinowernli.me/almanac/storage"

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

// New returns a new appender backed by the supplied storage.
func New(storage storage.Storage, discovery *discovery.Discovery) *mixer {
	return &mixer{storage: storage, discovery: discovery}
}

func (m *mixer) Search(ctx context.Context, request *pb_almanac.SearchRequest) (*pb_almanac.SearchResponse, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "search not implemented")
}
