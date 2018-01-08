package main

import (
	"encoding/json"
	"fmt"
	"testing"

	pb_almanac "dinowernli.me/almanac/proto"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

const (
	numAppenders   = 5 // The number of appenders in the system.
	appenderFanout = 2 // The number of appender each ingester talks to.

	grpcBasePort = 51000
)

var (
	testConf = &config{
		smallChunkMaxEntries: 10,
		smallChunkSpreadMs:   5000,
		smallChunkMaxAgeMs:   3000,
	}
)

func TestNoEntries(t *testing.T) {
	c := createTestCluster(t)
	defer c.stop()

	request := &pb_almanac.SearchRequest{Num: 200, Query: "foo"}
	response, err := c.mixer.Search(context.Background(), request)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(response.Entries))
}

func TestSearchesAppenders(t *testing.T) {
	c := createTestCluster(t)
	defer c.stop()

	// Add an entry containing "foo" to the first appender.
	append1, err := appendRequest("id1", "foo", 123)
	assert.NoError(t, err)

	_, err = c.appenders[0].Append(context.Background(), append1)
	assert.NoError(t, err)

	// Add a different entry containing "foo" to another appender.
	append2, err := appendRequest("id2", "foo", 567)
	assert.NoError(t, err)

	_, err = c.appenders[2].Append(context.Background(), append2)
	assert.NoError(t, err)

	// Add an entry which does not contain "foo".
	append3, err := appendRequest("id3", "baz", 789)
	assert.NoError(t, err)

	_, err = c.appenders[1].Append(context.Background(), append3)
	assert.NoError(t, err)

	// Make sure we get two hits when we search for foo.
	request := &pb_almanac.SearchRequest{Num: 200, Query: "foo"}
	response, err := c.mixer.Search(context.Background(), request)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(response.Entries))
}

func TestRoundTripThroughStorage(t *testing.T) {
	c := createTestCluster(t)
	defer c.stop()

	// Add multiple chunks worth of entries, try to make sure we have an
	// open chunk as well.
	numEntries := 10*testConf.smallChunkMaxEntries + 1
	for i := 0; i < numEntries; i++ {
		request, err := appendRequest(fmt.Sprintf("id-%d", i), "foo", 123)
		assert.NoError(t, err)

		_, err = c.appenders[0].Append(context.Background(), request)
		assert.NoError(t, err)
	}

	// Make sure all entries turn up.
	request := &pb_almanac.SearchRequest{Num: 200, Query: "foo"}
	response, err := c.mixer.Search(context.Background(), request)
	assert.NoError(t, err)
	assert.Equal(t, numEntries, len(response.Entries))
}

func TestDeduplicatesEntries(t *testing.T) {
	c := createTestCluster(t)
	defer c.stop()

	// Add an entry containing "foo" to multiple appenders.
	append1, err := appendRequest("id1", "foo", 123)
	assert.NoError(t, err)

	_, err = c.appenders[0].Append(context.Background(), append1)
	assert.NoError(t, err)

	_, err = c.appenders[1].Append(context.Background(), append1)
	assert.NoError(t, err)

	// Now make sure that a search for foo only returns the entry once.
	request := &pb_almanac.SearchRequest{Num: 200, Query: "foo"}
	response, err := c.mixer.Search(context.Background(), request)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(response.Entries))
}

func TestIngestsEntry(t *testing.T) {
	c := createTestCluster(t)
	defer c.stop()

	// Ship an entry with a predefined timestamp.
	ingestRequest1, err := newIngestRequest(&entry{Message: "foo", TimestampMs: 5000})
	assert.NoError(t, err)
	_, err = c.ingester.Ingest(context.Background(), ingestRequest1)
	assert.NoError(t, err)

	// Not timestamp for this entry, the ingester should infer it.
	ingestRequest2, err := newIngestRequest(&entry{Message: "foo"})
	assert.NoError(t, err)
	_, err = c.ingester.Ingest(context.Background(), ingestRequest2)
	assert.NoError(t, err)

	request := &pb_almanac.SearchRequest{Num: 200, Query: "foo"}
	response, err := c.mixer.Search(context.Background(), request)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(response.Entries))

	// Make sure we have the resulting entries sorted in ascending order by timestamp.
	assert.Equal(t, ingestRequest1.EntryJson, response.Entries[0].EntryJson)
	assert.Equal(t, int64(5000), response.Entries[0].TimestampMs)

	assert.Equal(t, ingestRequest2.EntryJson, response.Entries[1].EntryJson)
	assert.Equal(t, true, response.Entries[1].TimestampMs > 6000)
}

func TestQueryRange(t *testing.T) {
	c := createTestCluster(t)
	defer c.stop()

	ingestRequest1, err := newIngestRequest(&entry{Message: "foo", TimestampMs: 5000})
	assert.NoError(t, err)
	_, err = c.ingester.Ingest(context.Background(), ingestRequest1)
	assert.NoError(t, err)

	// Issue a query for a range which does not contain the value above.
	request := &pb_almanac.SearchRequest{Num: 200, Query: "foo", StartMs: 3000, EndMs: 4000}
	response, err := c.mixer.Search(context.Background(), request)
	assert.NoError(t, err)

	assert.Equal(t, 0, len(response.Entries))
}

func createTestCluster(t *testing.T) *localCluster {
	c, err := createCluster(logrus.New(), testConf, grpcBasePort, numAppenders, appenderFanout)
	assert.NoError(t, err)
	return c
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
