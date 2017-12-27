package appender

import (
	pb_almanac "dinowernli.me/almanac/proto"
	dc "dinowernli.me/almanac/service/discovery"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type Ingester struct {
}

func New(discovery *dc.Discovery) *Ingester {
	return &Ingester{}
}

func (i *Ingester) Ingest(ctx context.Context, request *pb_almanac.IngestRequest) (*pb_almanac.IngestResponse, error) {
	return &pb_almanac.IngestResponse{}, grpc.Errorf(codes.Unimplemented, "boo")
}
