package mixer

import (
	pb_almanac "dinowernli.me/almanac/proto"
	"dinowernli.me/almanac/storage"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type mixer struct {
	storage storage.Storage
}

// New returns a new appender backed by the supplied storage.
func New(storage storage.Storage) *mixer {
	return &mixer{storage: storage}
}

func (m *mixer) Search(ctx context.Context, request *pb_almanac.SearchRequest) (*pb_almanac.SearchResponse, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "search not implemented")
}
