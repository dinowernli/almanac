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

	startingPort = 51000
)

// fixture holds a test setup ready to use for testing.
type fixture struct {
	storage   st.Storage
	discovery *dc.Discovery
	mixer     *mx.Mixer
}

// entry represents a log entry as would be supplied by a user of the system.
type entry struct {
	timestamp_ms int
	message      string
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

func searchRequest(query string) (*pb_almanac.SearchRequest, error) {
	bleveRequest := bleve.NewSearchRequest(bleve.NewMatchQuery(query))
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
	port := startingPort
	storage := st.NewInMemoryStorage()
	appenders := []string{}
	for i := 0; i < numAppenders; i++ {
		appender, err := appender.New(fmt.Sprintf("appender-%d", i), storage, entriesPerChunk)
		if err != nil {
			return nil, fmt.Errorf("unable to create appender %d: %v", i, err)
		}

		address := fmt.Sprintf("localhost:%d", port)
		port++

		err = startAppenderServer(appender, address)
		if err != nil {
			return nil, fmt.Errorf("unable to start appender %d: %v", i, err)
		}

		appenders = append(appenders, address)
	}

	discovery, err := dc.New(appenders)
	if err != nil {
		return nil, fmt.Errorf("unable to create discovery: %v", err)
	}

	return &fixture{
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
