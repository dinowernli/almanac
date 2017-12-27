package ingester

import (
	"fmt"
	"log"

	pb_almanac "dinowernli.me/almanac/proto"
	dc "dinowernli.me/almanac/service/discovery"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Ingester is an implementation of the ingester service. It accepts log
// entries entering the system and fans them out to appenders.
type Ingester struct {
	discovery      *dc.Discovery
	appenderFanout int
}

// New returns a new Ingester backed by the supplied service discovery.
// appenderFanout specifies how many appenders this ingester tries to inform of
// a new log entry before declaring the entry ingested into the system.
func New(discovery *dc.Discovery, appenderFanout int) (*Ingester, error) {
	if appenderFanout < 1 {
		return nil, fmt.Errorf("appenderFanout must be at least 1")
	}

	return &Ingester{
		discovery:      discovery,
		appenderFanout: appenderFanout,
	}, nil
}

func (i *Ingester) Ingest(ctx context.Context, request *pb_almanac.IngestRequest) (*pb_almanac.IngestResponse, error) {
	appenders, err := i.selectAppenders()
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "unable to select appenders: %v", err)
	}

	// TODO(dino): Send an entry to every appender, return when done.
	log.Printf("sending to %d appenders\n", len(appenders))

	return &pb_almanac.IngestResponse{}, nil
}

func (i *Ingester) selectAppenders() ([]pb_almanac.AppenderClient, error) {
	allAppenders := i.discovery.ListAppenders()
	if i.appenderFanout < len(allAppenders) {
		return nil, fmt.Errorf("cannot select %d appender from a list of size %d", i.appenderFanout, len(allAppenders))
	}

	// TODO(dino): Consider remembering the appenders last used and trying to find them.
	// Shuffle the first time so that different ingesters use different subsets of appenders.

	return allAppenders[0:i.appenderFanout], nil
}
