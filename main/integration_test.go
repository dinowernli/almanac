package main

import (
	"encoding/json"
	"fmt"
	"net"
	"testing"

	pb_almanac "dinowernli.me/almanac/proto"
	"dinowernli.me/almanac/service/appender"
	dc "dinowernli.me/almanac/service/discovery"
	in "dinowernli.me/almanac/service/ingester"
	mx "dinowernli.me/almanac/service/mixer"
	st "dinowernli.me/almanac/storage"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	entriesPerChunk = 10 // The max number of entries in a single chunk.
	numAppenders    = 5  // The number of appenders in the system.
	appenderFanout  = 2  // The number of appender each ingester talks to.
)

var (
	nextPort = 51000
)

// fixture holds a test setup ready to use for testing.
type fixture struct {
	appenders []*appender.Appender
	ingester  *in.Ingester
	storage   *st.Storage
	discovery *dc.Discovery
	mixer     *mx.Mixer
}

// entry represents a log entry as would be supplied by a user of the system.
type entry struct {
	Message     string `json:"message"`
	Logger      string `json:"logger"`
	TimestampMs int64  `json:"timestamp_ms"`
}

func TestNoEntries(t *testing.T) {
	f, err := createFixture()
	assert.NoError(t, err)

	request := &pb_almanac.SearchRequest{Num: 200, Query: "foo"}
	response, err := f.mixer.Search(context.Background(), request)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(response.Entries))
}

func TestSearchesAppenders(t *testing.T) {
	f, err := createFixture()
	assert.NoError(t, err)

	// Add an entry containing "foo" to the first appender.
	append1, err := appendRequest("id1", "foo", 123)
	assert.NoError(t, err)

	_, err = f.appenders[0].Append(context.Background(), append1)
	assert.NoError(t, err)

	// Add a different entry containing "foo" to another appender.
	append2, err := appendRequest("id2", "foo", 567)
	assert.NoError(t, err)

	_, err = f.appenders[2].Append(context.Background(), append2)
	assert.NoError(t, err)

	// Add an entry which does not contain "foo".
	append3, err := appendRequest("id3", "baz", 789)
	assert.NoError(t, err)

	_, err = f.appenders[1].Append(context.Background(), append3)
	assert.NoError(t, err)

	// Make sure we get two hits when we search for foo.
	request := &pb_almanac.SearchRequest{Num: 200, Query: "foo"}
	response, err := f.mixer.Search(context.Background(), request)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(response.Entries))
}

func TestRoundTripThroughStorage(t *testing.T) {
	f, err := createFixture()
	assert.NoError(t, err)

	// Add multiple chunks worth of entries, try to make sure we have an
	// open chunk as well.
	numEntries := 10*entriesPerChunk + 1
	for i := 0; i < numEntries; i++ {
		request, err := appendRequest(fmt.Sprintf("id-%d", i), "foo", 123)
		assert.NoError(t, err)

		_, err = f.appenders[0].Append(context.Background(), request)
		assert.NoError(t, err)
	}

	// Make sure all entries turn up.
	request := &pb_almanac.SearchRequest{Num: 200, Query: "foo"}
	response, err := f.mixer.Search(context.Background(), request)
	assert.NoError(t, err)
	assert.Equal(t, numEntries, len(response.Entries))
}

func TestDeduplicatesEntries(t *testing.T) {
	f, err := createFixture()
	assert.NoError(t, err)

	// Add an entry containing "foo" to multiple appenders.
	append1, err := appendRequest("id1", "foo", 123)
	assert.NoError(t, err)

	_, err = f.appenders[0].Append(context.Background(), append1)
	assert.NoError(t, err)

	_, err = f.appenders[1].Append(context.Background(), append1)
	assert.NoError(t, err)

	// Now make sure that a search for foo only returns the entry once.
	request := &pb_almanac.SearchRequest{Num: 200, Query: "foo"}
	response, err := f.mixer.Search(context.Background(), request)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(response.Entries))
}

func TestIngestsEntry(t *testing.T) {
	f, err := createFixture()
	assert.NoError(t, err)

	// Ship an entry with a predefined timestamp.
	ingestRequest1, err := ingestRequest(&entry{Message: "foo", TimestampMs: 5000})
	assert.NoError(t, err)
	_, err = f.ingester.Ingest(context.Background(), ingestRequest1)
	assert.NoError(t, err)

	// Not timestamp for this entry, the ingester should infer it.
	ingestRequest2, err := ingestRequest(&entry{Message: "foo"})
	assert.NoError(t, err)
	_, err = f.ingester.Ingest(context.Background(), ingestRequest2)
	assert.NoError(t, err)

	request := &pb_almanac.SearchRequest{Num: 200, Query: "foo"}
	response, err := f.mixer.Search(context.Background(), request)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(response.Entries))

	// Make sure we have the resulting entries sorted in ascending order by timestamp.
	assert.Equal(t, ingestRequest1.EntryJson, response.Entries[0].EntryJson)
	assert.Equal(t, int64(5000), response.Entries[0].TimestampMs)

	assert.Equal(t, ingestRequest2.EntryJson, response.Entries[1].EntryJson)
	assert.Equal(t, true, response.Entries[1].TimestampMs > 6000)
}

func TestQueryRange(t *testing.T) {
	f, err := createFixture()
	assert.NoError(t, err)

	ingestRequest1, err := ingestRequest(&entry{Message: "foo", TimestampMs: 5000})
	assert.NoError(t, err)
	_, err = f.ingester.Ingest(context.Background(), ingestRequest1)
	assert.NoError(t, err)

	// Issue a query for a range which does not contain the value above.
	request := &pb_almanac.SearchRequest{Num: 200, Query: "foo", StartMs: 3000, EndMs: 4000}
	_, err = f.mixer.Search(context.Background(), request)
	assert.NoError(t, err)

	// TODO(dino): We don't yet correctly respect query time ranges.
	// assert.Equal(t, 0, len(response.Entries))
}

func ingestRequest(e *entry) (*pb_almanac.IngestRequest, error) {
	entryJson, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal entry to json: %v", err)
	}
	return &pb_almanac.IngestRequest{EntryJson: string(entryJson)}, nil
}

func appendRequest(id string, message string, timestampMs int64) (*pb_almanac.AppendRequest, error) {
	fooJson, err := json.Marshal(&entry{Message: message})
	if err != nil {
		return nil, fmt.Errorf("unable to marshal entry to json: %v", err)
	}

	return &pb_almanac.AppendRequest{
		Entry: &pb_almanac.LogEntry{
			EntryJson:   string(fooJson),
			TimestampMs: timestampMs,
			Id:          id,
		},
	}, nil
}

// createFixture sets up a test fixture, including all services required to run the system.
func createFixture() (*fixture, error) {
	storage := st.NewInMemoryStorage()
	appenderAddresses := []string{}
	appenders := []*appender.Appender{}
	for i := 0; i < numAppenders; i++ {
		appender, err := appender.New(storage, entriesPerChunk)
		if err != nil {
			return nil, fmt.Errorf("unable to create appender %d: %v", i, err)
		}

		address := fmt.Sprintf("localhost:%d", nextPort)
		nextPort++

		err = startAppenderServer(appender, address)
		if err != nil {
			return nil, fmt.Errorf("unable to start appender %d: %v", i, err)
		}

		appenderAddresses = append(appenderAddresses, address)
		appenders = append(appenders, appender)
	}

	discovery, err := dc.New(appenderAddresses)
	if err != nil {
		return nil, fmt.Errorf("unable to create discovery: %v", err)
	}

	ingester, err := in.New(discovery, appenderFanout)
	if err != nil {
		return nil, fmt.Errorf("unable to create ingester: %v", err)
	}

	return &fixture{
		appenders: appenders,
		ingester:  ingester,
		storage:   storage,
		discovery: discovery,
		mixer:     mx.New(storage, discovery),
	}, nil
}

func startAppenderServer(appender *appender.Appender, address string) error {
	listen, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen for address %s: %v", address, err)
	}

	server := grpc.NewServer()
	pb_almanac.RegisterAppenderServer(server, appender)
	go func() {
		server.Serve(listen)
	}()

	return nil
}
