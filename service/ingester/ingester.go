package ingester

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	pb_almanac "dinowernli.me/almanac/proto"
	dc "dinowernli.me/almanac/service/discovery"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const (
	timestampField = "timestamp_ms"
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
	entry, err := extractEntry(request.EntryJson)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "unable to extract log entry from json: %v", err)
	}

	appenders, err := i.selectAppenders()
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "unable to select appenders: %v", err)
	}

	// TODO(dino): Send an entry to every appender, return when done.
	log.Printf("sending to %d appenders, %v\n", len(appenders), entry)

	return &pb_almanac.IngestResponse{}, nil
}

// selectAppenders returns the appenders to whom the current log entry must be
// sent to.
func (i *Ingester) selectAppenders() ([]pb_almanac.AppenderClient, error) {
	allAppenders := i.discovery.ListAppenders()
	if i.appenderFanout < len(allAppenders) {
		return nil, fmt.Errorf("cannot select %d appender from a list of size %d", i.appenderFanout, len(allAppenders))
	}

	// TODO(dino): Consider remembering the appenders last used and trying to find them.
	// Shuffle the first time so that different ingesters use different subsets of appenders.

	return allAppenders[0:i.appenderFanout], nil
}

// extractEntry takes an incoming string and construct a LogEntry proto from
// it. This can fail if the incoming entry is not valid json, or if the json
// is otherwise malformed.
func extractEntry(rawJson string) (*pb_almanac.LogEntry, error) {
	var rawEntry map[string]*json.RawMessage
	err := json.Unmarshal([]byte(rawJson), &rawEntry)
	if err != nil {
		return nil, fmt.Errorf("unable to parse raw entry")
	}

	var timestampMs int64

	// Attempt to extract a timestamp.
	foundTimestamp := false
	value, ok := rawEntry[timestampField]
	if ok {
		err := json.Unmarshal(*value, &timestampMs)
		if err != nil {
			return nil, fmt.Errorf("could not parse value for timestamp: %v", err)
		}
		foundTimestamp = true
	}

	// If no timestamp was provided, just use "now".
	if !foundTimestamp {
		timestampMs = time.Now().Unix()
	}

	// TODO(dino): Actually compue an id for the entry.
	id := "boo"
	return &pb_almanac.LogEntry{
		EntryJson:   rawJson,
		TimestampMs: timestampMs,
		Id:          id,
	}, nil
}
