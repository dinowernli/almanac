package ingester

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	pb_almanac "dinowernli.me/almanac/proto"
	dc "dinowernli.me/almanac/service/discovery"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const (
	timestampField = "timestamp_ms"
	nanosPerMilli  = 1000000
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
	// Parse the incoming raw log entry, extracting some structure.
	entry, err := extractEntry(request.EntryJson)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "unable to extract log entry from json: %v", err)
	}

	// Send an append request to a select bunch of appenders.
	appenders, err := i.selectAppenders()
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "unable to select appenders: %v", err)
	}

	appendRequest := &pb_almanac.AppendRequest{Entry: entry}
	resultChan := make(chan error, len(appenders))
	for _, appender := range appenders {
		go func() {
			_, err := appender.Append(ctx, appendRequest)
			resultChan <- err
		}()
	}

	// TODO(dino): be smarter about how many append calls need to have succeeded. For now,
	// we error out if any of the append calls fail.
	for i := 0; i < len(appenders); i++ {
		err := <-resultChan
		if err != nil {
			return nil, grpc.Errorf(codes.Internal, "unable to send append request: %v", err)
		}
	}

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
		timestampMs = time.Now().UnixNano() / nanosPerMilli
	}

	return &pb_almanac.LogEntry{
		EntryJson:   rawJson,
		TimestampMs: timestampMs,
		Id:          newEntryId(timestampMs),
	}, nil
}

// newEntryId returns a string id for an entry with the given timestamp. The
// ids have the property that sorting them lexicographically orders them by
// timestamp, but that no two different entries ever end up with the same id.
func newEntryId(timestampMs int64) string {
	return fmt.Sprintf("%d-%s", timestampMs, randomString(3))
}

// TODO(dino): Deduplicate these methods with appender.go.
// randomString produces a random string of lower case letters.
func randomString(num int) string {
	bytes := make([]byte, num)
	for i := 0; i < num; i++ {
		bytes[i] = byte(randomInt(97, 122)) // lowercase letters.
	}
	return string(bytes)
}

func randomInt(min int, max int) int {
	return min + rand.Intn(max-min)
}
