package appender

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"dinowernli.me/almanac/index"
	pb_almanac "dinowernli.me/almanac/proto"
	"dinowernli.me/almanac/storage"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// appender keeps track of a single open chunk under construction at a time and
// periodically decides that the open chunk is complete, at which point the
// appender writes the chunk to storage. The appender also knows how to answer
// search requests for the currently open chunk.
type appender struct {
	entries     []*pb_almanac.LogEntry
	index       *index.Index
	chunkId     int
	appendMutex *sync.Mutex

	appenderId         string
	storage            storage.Storage
	maxEntriesPerChunk int
}

// New returns a new appender backed by the supplied storage.
func New(appenderId string, storage storage.Storage, maxEntriesPerChunk int) (*appender, error) {
	if maxEntriesPerChunk < 1 {
		return nil, fmt.Errorf("max entries per chunk must be greater than 0, but got %d", maxEntriesPerChunk)
	}

	index, err := index.NewIndex()
	if err != nil {
		return nil, fmt.Errorf("unable to create index: %v", err)
	}

	return &appender{
		index:              index,
		appendMutex:        &sync.Mutex{},
		storage:            storage,
		maxEntriesPerChunk: maxEntriesPerChunk,
	}, nil
}

func (a *appender) Search(ctx context.Context, request *pb_almanac.SearchRequest) (*pb_almanac.SearchResponse, error) {
	a.appendMutex.Lock()
	defer a.appendMutex.Unlock()

	response, err := a.index.Search(ctx, request)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "unable to search: %v", err)
	}
	return response, nil
}

func (a *appender) Append(ctx context.Context, request *pb_almanac.AppendRequest) (*pb_almanac.AppendResponse, error) {
	// Perform some validation of the request.
	logEntry := request.GetEntry()
	if logEntry == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "no entry supplied")
	}

	entryId := logEntry.GetId()
	if entryId == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "no id supplied")
	}

	// Update the data structures corresponding to the open chunk.
	a.appendMutex.Lock()
	defer a.appendMutex.Unlock()

	var rawEntry interface{}
	err := json.Unmarshal([]byte(logEntry.EntryJson), &rawEntry)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "unable to parse raw json: %v", err)
	}

	err = a.index.Index(entryId, rawEntry)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "unable to index raw json entry: %v", err)
	}

	a.entries = append(a.entries, request.Entry)

	// If the open chunk has grown enough, close it up and start a new one.
	if len(a.entries) >= a.maxEntriesPerChunk {
		err := a.storeChunk()
		if err != nil {
			return nil, grpc.Errorf(codes.Internal, "unable to store chunk: %v", err)
		}

		index, err := index.NewIndex()
		if err != nil {
			// TODO(dino): If this happens, stop responding to requests.
			return nil, grpc.Errorf(codes.Internal, "unable to create new index: %v", err)
		}
		a.index = index
		a.entries = []*pb_almanac.LogEntry{}
	}

	return &pb_almanac.AppendResponse{}, nil
}

// storeChunk takes the currently open chunk, persists it to storage, and starts
// a new open chunk.
func (a *appender) storeChunk() error {
	indexProto, err := index.Serialize(a.index)
	if err != nil {
		return fmt.Errorf("unable to serialize index: %v", err)
	}

	chunkProto := &pb_almanac.Chunk{
		Entries: a.entries,
		Index:   indexProto,
	}

	bytes, err := proto.Marshal(chunkProto)
	if err != nil {
		return fmt.Errorf("unable to marshal chunk proto: %v", err)
	}
	chunkName := a.nextChunkName()
	a.storage.Write(chunkName, bytes)
	log.Printf("wrote chunk: %s", chunkName)

	return nil
}

func (a *appender) nextChunkName() string {
	result := fmt.Sprintf("chunk-%d-%s", a.chunkId, a.appenderId)
	a.chunkId++
	return result
}
