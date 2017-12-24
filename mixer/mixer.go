package mixer

import (
	"fmt"

	"dinowernli.me/almanac/discovery"
	pb_almanac "dinowernli.me/almanac/proto"
	"dinowernli.me/almanac/storage"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type Mixer struct {
	storage   *storage.Storage
	discovery *discovery.Discovery
}

// New returns a new mixer backed by the supplied storage.
func New(storage *storage.Storage, discovery *discovery.Discovery) *Mixer {
	return &Mixer{storage: storage, discovery: discovery}
}

func (m *Mixer) searchChunk(chunkId string, query string, num int32, resultChan chan *partialResult) {
	result := &partialResult{}
	chunk, err := m.storage.LoadChunk(chunkId)
	if err != nil {
		result.err = fmt.Errorf("unable to load chunk %s: %v\n", chunkId, err)
		resultChan <- result
	}
	result.chunk = chunk

	entries, err := chunk.Search(query, num)
	if err != nil {
		result.err = fmt.Errorf("unable to perform search on chunk %s: %v\n", chunkId, err)
		resultChan <- result
	}

	result.entries = entries
	resultChan <- result
}

type partialResult struct {
	chunk   *storage.Chunk
	entries []*pb_almanac.LogEntry
	err     error
}

func (m *Mixer) searchAppender(ctx context.Context, appender pb_almanac.AppenderClient, request *pb_almanac.SearchRequest, resultChan chan *partialResult) {
	response, err := appender.Search(ctx, request)
	if err != nil {
		resultChan <- &partialResult{err: err}
		return
	}
	resultChan <- &partialResult{entries: response.Entries}
}

func (m *Mixer) Search(ctx context.Context, request *pb_almanac.SearchRequest) (*pb_almanac.SearchResponse, error) {
	// Do some prep for the parallel searches.
	appenders := m.discovery.ListAppenders()
	chunkIds, err := m.storage.ListChunks(request.StartMs, request.EndMs)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "unable to list chunks: %v", err)
	}

	// Execute all the searches in parallel.
	numSubRequests := len(appenders) + len(chunkIds)
	resultChan := make(chan *partialResult, numSubRequests)
	for _, chunkId := range chunkIds {
		go m.searchChunk(chunkId, request.Query, request.Num, resultChan)
	}
	for _, appender := range appenders {
		go m.searchAppender(ctx, appender, request, resultChan)
	}

	// Drain the channel and collect all the entries.
	allEntries := []*pb_almanac.LogEntry{}
	err = nil
	for i := 0; i < numSubRequests; i++ {
		result := <-resultChan
		if result.chunk != nil {
			result.chunk.Close()
		}

		if result.err == nil {
			for _, e := range result.entries {
				allEntries = append(allEntries, e)
			}
		} else {
			err = result.err
		}
	}

	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "error executing search: %v", err)
	}

	// TODO(dino): Sort and truncate. For now, just return everything.
	return &pb_almanac.SearchResponse{Entries: allEntries}, nil
}
