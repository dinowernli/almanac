package main

import (
	"encoding/json"
	"fmt"
	"net"
	"testing"

	"dinowernli.me/almanac/appender"
	dc "dinowernli.me/almanac/discovery"
	mx "dinowernli.me/almanac/mixer"
	pb_almanac "dinowernli.me/almanac/proto"
	st "dinowernli.me/almanac/storage"

	"github.com/blevesearch/bleve"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	entriesPerChunk = 10
	numAppenders    = 5
)

var (
	nextPort = 51000
)

// fixture holds a test setup ready to use for testing.
type fixture struct {
	appenders []*appender.Appender
	storage   st.Storage
	discovery *dc.Discovery
	mixer     *mx.Mixer
}

// entry represents a log entry as would be supplied by a user of the system.
type entry struct {
	Message string
}

func TestNoEntries(t *testing.T) {
	f, err := createFixture()
	assert.NoError(t, err)

	request, err := searchRequest("foo")
	assert.NoError(t, err)

	response, err := f.mixer.Search(context.Background(), request)
	assert.NoError(t, err)

	bleveResponse, err := unpackResponse(response)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(bleveResponse.Hits))
}

func TestSearchesAppenders(t *testing.T) {
	f, err := createFixture()
	assert.NoError(t, err)

	// Add an entry containing "foo" to the first appender.
	append1, err := appendRequest("entry1", "foo", 123)
	assert.NoError(t, err)

	_, err = f.appenders[0].Append(context.Background(), append1)
	assert.NoError(t, err)

	// Add a different entry containing "foo" to another appender.
	append2, err := appendRequest("entry2", "foo", 567)
	assert.NoError(t, err)

	_, err = f.appenders[2].Append(context.Background(), append2)
	assert.NoError(t, err)

	// Make sure we get two hits when we search for foo.
	append3, err := appendRequest("entry3", "baz", 789)
	assert.NoError(t, err)

	_, err = f.appenders[1].Append(context.Background(), append3)
	assert.NoError(t, err)

	// Now perform some searches.
	request, err := searchRequest("foo")
	assert.NoError(t, err)

	response, err := f.mixer.Search(context.Background(), request)
	assert.NoError(t, err)

	bleveResponse, err := unpackResponse(response)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(bleveResponse.Hits))
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
	request, err := searchRequest("foo")
	assert.NoError(t, err)

	response, err := f.mixer.Search(context.Background(), request)
	assert.NoError(t, err)

	bleveResponse, err := unpackResponse(response)
	assert.NoError(t, err)
	assert.Equal(t, numEntries, len(bleveResponse.Hits))
}

func TestDeduplicatedEntries(t *testing.T) {
	f, err := createFixture()
	assert.NoError(t, err)

	// Add an entry containing "foo" to multiple appenders.
	append1, err := appendRequest("entry1", "foo", 123)
	assert.NoError(t, err)

	_, err = f.appenders[0].Append(context.Background(), append1)
	assert.NoError(t, err)

	_, err = f.appenders[1].Append(context.Background(), append1)
	assert.NoError(t, err)

	// Now perform some searches.
	request, err := searchRequest("foo")
	assert.NoError(t, err)

	response, err := f.mixer.Search(context.Background(), request)
	assert.NoError(t, err)

	_, err = unpackResponse(response)
	assert.NoError(t, err)

	// TODO(dino): Teach bleve how to dedupe docs based on their id.
	// assert.Equal(t, 1, len(bleveResponse.Hits))
}

func appendRequest(id string, message string, timestampMs int64) (*pb_almanac.AppendRequest, error) {
	fooJson, err := json.Marshal(&entry{message})
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

func searchRequest(query string) (*pb_almanac.SearchRequest, error) {
	bleveRequest := bleve.NewSearchRequestOptions(
		bleve.NewMatchQuery(query),
		200, /* size */
		0,   /* from */
		false /* explain */)
	bleveBytes, err := json.Marshal(bleveRequest)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal bleve request: %v", err)
	}
	return &pb_almanac.SearchRequest{BleveRequestBytes: bleveBytes}, nil
}

func unpackResponse(response *pb_almanac.SearchResponse) (*bleve.SearchResult, error) {
	result := &bleve.SearchResult{}
	err := json.Unmarshal(response.BleveResponseBytes, result)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal bleve response: %v", err)
	}
	return result, nil
}

// createFixture sets up a test fixture, including all services required to run the system.
func createFixture() (*fixture, error) {
	storage := st.NewInMemoryStorage()
	appenderAddresses := []string{}
	appenders := []*appender.Appender{}
	for i := 0; i < numAppenders; i++ {
		appender, err := appender.New(fmt.Sprintf("appender-%d", i), storage, entriesPerChunk)
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

	return &fixture{
		appenders: appenders,
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
